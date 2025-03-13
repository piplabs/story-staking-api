package indexer

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

var _ Indexer = (*ELBlockIndexer)(nil)

type ELBlockIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	ethClient *ethclient.Client
}

func NewELBlockIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, rpcEndpoint string) (*ELBlockIndexer, error) {
	ethClient, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return nil, err
	}

	return &ELBlockIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		ethClient: ethClient,
	}, nil
}

func (e *ELBlockIndexer) Name() string {
	return "el_block"
}

func (e *ELBlockIndexer) Run() {
	log.Info().Str("indexer", e.Name()).Msg("Start indexing")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			indexPoint, err := db.GetIndexPoint(e.dbOperator, e.Name())
			if err != nil {
				log.Error().Err(err).Str("indexer", e.Name()).Msg("get index point failed")
				continue
			}

			latestBlkNum, err := e.ethClient.BlockNumber(e.ctx)
			if err != nil {
				log.Error().Err(err).Str("indexer", e.Name()).Msg("get latest el block failed")
				continue
			}

			if indexPoint.BlockHeight+10 > int64(latestBlkNum) {
				continue
			}

			if err := e.index(indexPoint.BlockHeight+1, int64(latestBlkNum)); err != nil {
				log.Error().Err(err).
					Str("indexer", e.Name()).
					Int64("from", indexPoint.BlockHeight+1).
					Uint64("to", latestBlkNum).
					Msg("index el block failed")
			}
		}
	}
}

func (e *ELBlockIndexer) index(from, to int64) error {
	var elBlocks []*db.ELBlock

	for i := from; i <= to; i++ {
		blk, err := e.ethClient.BlockByNumber(e.ctx, big.NewInt(i))
		if err != nil {
			return err
		}

		elBlocks = append(elBlocks, &db.ELBlock{
			Height:   i,
			Hash:     blk.Hash().String(),
			GasUsed:  blk.GasUsed(),
			GasLimit: blk.GasLimit(),
			Time:     time.Unix(int64(blk.Time()), 0),
		})

		if len(elBlocks) > 100 {
			if err := db.BatchCreateELBlocks(e.dbOperator, e.Name(), elBlocks, i); err != nil {
				return err
			}

			elBlocks = make([]*db.ELBlock, 0)
		}
	}

	// Handle remaining entries, even if there are no entries, we also need to update the index point.
	if err := db.BatchCreateELBlocks(e.dbOperator, e.Name(), elBlocks, to); err != nil {
		return err
	}

	return nil
}
