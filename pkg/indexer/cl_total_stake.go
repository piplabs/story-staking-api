package indexer

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

const (
	genesisStakeAmount = 8000000 // gwei

	genesisTime = "2025-01-19T15:00:00.00000000Z"
	timeLayout  = "2006-01-02T15:04:05.999999999Z"

	stakeEventType   = "delegate_success"
	unstakeEventType = "undelegate_success"
)

var _ Indexer = (*CLTotalStakeIndexer)(nil)

type CLStakeChange struct {
	BlockHeight       int64
	StakeChangeAmount int64
}

type CLTotalStakeIndexer struct {
	ctx context.Context

	dbOperator *gorm.DB

	cometClient            *comethttp.HTTP
	latestTotalStakeAmount int64
}

func NewCLTotalStakeIndexer(ctx context.Context, dbOperator *gorm.DB, rpcEndpoint string) (*CLTotalStakeIndexer, error) {
	cometClient, err := comethttp.New(rpcEndpoint, "")
	if err != nil {
		return nil, err
	}

	c := &CLTotalStakeIndexer{
		ctx: ctx,

		dbOperator: dbOperator,

		cometClient: cometClient,
	}

	if err := c.init(); err != nil {
		return nil, fmt.Errorf("init indexer %s failed: %w", c.Name(), err)
	}

	return c, nil
}

func (c *CLTotalStakeIndexer) Name() string {
	return "cl_total_stake"
}

func (c *CLTotalStakeIndexer) Run() {
	log.Info().Str("indexer", c.Name()).Msg("Start indexing")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			indexPoint, err := db.GetIndexPoint(c.dbOperator, c.Name())
			if err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("get index point failed")
				continue
			}

			latestBlk, err := c.cometClient.Block(c.ctx, nil)
			if err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("get latest cl block failed")
				continue
			}

			if indexPoint.BlockHeight+10 > latestBlk.Block.Height {
				continue
			}

			from, to := indexPoint.BlockHeight+1, latestBlk.Block.Height
			if err := c.index(from, to); err != nil {
				log.Error().Err(err).
					Int64("from", from).
					Int64("to", to).
					Str("indexer", c.Name()).
					Msg("index cl total stake history failed")
			}
		}
	}
}

func (c *CLTotalStakeIndexer) init() error {
	gt, err := time.Parse(timeLayout, genesisTime)
	if err != nil {
		return fmt.Errorf("parse genesis time failed: %w", err)
	}

	if err := db.InsertCLTotalStake(c.dbOperator, c.Name(), &db.CLTotalStake{
		UpdateAt:         gt.Unix(),
		TotalStakeAmount: genesisStakeAmount,
	}); err != nil {
		return fmt.Errorf("upsert cl total stake failed: %w", err)
	}

	row, err := db.GetLatestCLTotalStake(c.dbOperator)
	if err != nil {
		return fmt.Errorf("get latest cl total stake failed: %w", err)
	}

	c.latestTotalStakeAmount = row.TotalStakeAmount

	return nil
}

func (c *CLTotalStakeIndexer) index(from, to int64) error {
	start := from
	for start <= to {
		end := min(start+100, to)

		if err := c.applyStakeChanges(start, end); err != nil {
			return err
		}

		start = end + 1
	}

	return nil
}

func (c *CLTotalStakeIndexer) applyStakeChanges(from, to int64) error {
	// Get all stake changes.
	stakeChanges, err := c.getCLStakeChanges(from, to)
	if err != nil {
		return fmt.Errorf("get cl stake changes failed: %w", err)
	}
	// Get all block height & time of the stake changes.
	clBlockHeightSet := make(map[int64]struct{})
	for _, sc := range stakeChanges {
		clBlockHeightSet[sc.BlockHeight] = struct{}{}
	}

	clBlockHeights := make([]int64, 0)
	for height := range clBlockHeightSet {
		clBlockHeights = append(clBlockHeights, height)
	}

	clBlocks, err := db.GetCLBlocks(c.dbOperator, clBlockHeights)
	if err != nil {
		return fmt.Errorf("get cl blocks failed: %w", err)
	}

	clBlockToTimestamp := make(map[int64]time.Time)
	for _, clBlock := range clBlocks {
		clBlockToTimestamp[clBlock.Height] = clBlock.Time
	}

	// Apply stake changes
	rows := make([]*db.CLTotalStake, 0)
	for _, sc := range stakeChanges {
		c.latestTotalStakeAmount += sc.StakeChangeAmount

		blockTime, ok := clBlockToTimestamp[sc.BlockHeight]
		if !ok {
			return fmt.Errorf("unexcepted getting block %d timestamp failed", sc.BlockHeight)
		}

		rows = append(rows, &db.CLTotalStake{
			UpdateAt:         blockTime.Unix(),
			TotalStakeAmount: c.latestTotalStakeAmount,
		})
	}

	if err := db.BatchUpsertCLTotalStakes(c.dbOperator, c.Name(), rows, to); err != nil {
		return fmt.Errorf("batch upsert cl total stakes failed: %w", err)
	}

	return nil
}

func (c *CLTotalStakeIndexer) getCLStakeChanges(from, to int64) ([]*CLStakeChange, error) {
	stakeChanges := make(map[int64]int64)

	for blkno := from; blkno <= to; blkno++ {
		blockResults, err := c.cometClient.BlockResults(c.ctx, &blkno)
		if err != nil {
			return nil, err
		}

		blockEvents := make([]abcitypes.Event, 0)
		for _, tr := range blockResults.TxsResults {
			blockEvents = append(blockEvents, tr.Events...)
		}
		blockEvents = append(blockEvents, blockResults.FinalizeBlockEvents...)

		for _, e := range blockEvents {
			if !(e.Type == stakeEventType || e.Type == unstakeEventType) {
				continue
			}

			attrMap := attrArray2Map(e.Attributes)

			if attrMap[AttributeKeyAmount] == "" {
				// singularity period
				continue
			}

			amount, err := strconv.ParseInt(attrMap[AttributeKeyAmount], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse amount %s failed: %w", attrMap[AttributeKeyAmount], err)
			}
			if e.Type == unstakeEventType {
				amount *= -1
			}

			if _, ok := stakeChanges[blkno]; !ok {
				stakeChanges[blkno] = amount
			} else {
				stakeChanges[blkno] += amount
			}
		}
	}

	scs := make([]*CLStakeChange, 0, len(stakeChanges))
	for blkno, amt := range stakeChanges {
		scs = append(scs, &CLStakeChange{
			BlockHeight:       blkno,
			StakeChangeAmount: amt,
		})
	}

	sort.Slice(scs, func(i, j int) bool {
		return scs[i].BlockHeight < scs[j].BlockHeight
	})

	return scs, nil
}
