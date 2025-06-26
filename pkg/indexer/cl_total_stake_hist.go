package indexer

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

const (
	genesisBlockHeight = 1
)

var _ Indexer = (*CLTotalStakeHistIndexer)(nil)

type CLTotalStakeHistIndexer struct {
	ctx context.Context

	dbOperator *gorm.DB
}

func NewCLTotalStakeHistIndexer(ctx context.Context, dbOperator *gorm.DB) (*CLTotalStakeHistIndexer, error) {
	c := &CLTotalStakeHistIndexer{
		ctx: ctx,

		dbOperator: dbOperator,
	}

	if err := c.init(); err != nil {
		return nil, fmt.Errorf("init indexer %s failed: %w", c.Name(), err)
	}

	return c, nil
}

func (c *CLTotalStakeHistIndexer) Name() string {
	return "cl_total_stake_hist"
}

func (c *CLTotalStakeHistIndexer) Run() {
	log.Info().Str("indexer", c.Name()).Msg("Start indexing")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.index(); err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("index cl total stake hist failed")
			}
		}
	}
}

func (c *CLTotalStakeHistIndexer) init() error {
	gt, err := time.Parse(timeLayout, genesisTime)
	if err != nil {
		return fmt.Errorf("parse genesis time failed: %w", err)
	}

	if err := db.UpsertCLGenesisTotalStakeHist(c.dbOperator, c.Name(), &db.CLTotalStakeHist{
		TotalStakeAmount: genesisStakeAmount,
		UpdatedAtBlock:   genesisBlockHeight,
		UpdatedAtTime:    gt.Unix(),
	}); err != nil {
		return fmt.Errorf("upsert genesis cl total stake failed: %w", err)
	}

	return nil
}

func (c *CLTotalStakeHistIndexer) index() error {
	latestHist, err := db.GetLatestCLTotalStakeHist(c.dbOperator)
	if err != nil {
		return fmt.Errorf("get latest cl total stake hist failed: %w", err)
	}

	events, err := db.GetSuccessfulCLStakingEventsAfter(c.dbOperator, []string{TypeStake, TypeUnstake}, latestHist.UpdatedAtBlock)
	if err != nil {
		return fmt.Errorf("get successful cl staking events failed: %w", err)
	} else if len(events) == 0 {
		return nil
	}

	blk2StakeChange, blk2BlockTime := make(map[int64]int64), make(map[int64]int64)
	for _, event := range events {
		var amount int64
		if event.Amount == "" {
			// special bypass for aeneid
			amount = 0
		} else {
			amt, err := strconv.ParseInt(event.Amount, 10, 64)
			if err != nil {
				return fmt.Errorf("parse amount %s failed: %w", event.Amount, err)
			}
			amount = amt
		}

		switch event.EventType {
		case TypeStake:
			blk2StakeChange[event.BlockHeight] += amount
		case TypeUnstake:
			blk2StakeChange[event.BlockHeight] -= amount
		}

		blk2BlockTime[event.BlockHeight] = event.BlockTime.Unix()
	}

	type stakeChange struct {
		BlockHeight       int64
		BlockTime         int64
		StakeChangeAmount int64
	}

	scs := make([]*stakeChange, 0)
	for blkno, amt := range blk2StakeChange {
		scs = append(scs, &stakeChange{
			BlockHeight:       blkno,
			BlockTime:         blk2BlockTime[blkno],
			StakeChangeAmount: amt,
		})
	}

	sort.Slice(scs, func(i, j int) bool {
		return scs[i].BlockHeight < scs[j].BlockHeight
	})

	totalStakeAmount := latestHist.TotalStakeAmount
	clTotalStakeHists := make([]*db.CLTotalStakeHist, 0)
	for _, sc := range scs {
		totalStakeAmount += sc.StakeChangeAmount
		clTotalStakeHists = append(clTotalStakeHists, &db.CLTotalStakeHist{
			TotalStakeAmount: totalStakeAmount,
			UpdatedAtBlock:   sc.BlockHeight,
			UpdatedAtTime:    sc.BlockTime,
		})
	}

	return db.BatchUpsertCLTotalStakeHists(c.dbOperator, c.Name(), clTotalStakeHists)
}
