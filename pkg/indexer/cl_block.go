package indexer

import (
	"context"
	"time"

	lightprovider "github.com/cometbft/cometbft/light/provider"
	lighthttp "github.com/cometbft/cometbft/light/provider/http"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-indexer/db"
)

var _ Indexer = (*CLBlockIndexer)(nil)

type CLBlockIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	lightCometClient lightprovider.Provider
}

func NewCLBlockIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, chainID, rpcEndpoint string) (*CLBlockIndexer, error) {
	lightCometClient, err := lighthttp.New(chainID, rpcEndpoint)
	if err != nil {
		return nil, err
	}

	return &CLBlockIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		lightCometClient: lightCometClient,
	}, nil
}

func (c *CLBlockIndexer) Name() string {
	return "cl_block"
}

func (c *CLBlockIndexer) Run() {
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

			latestBlk, err := c.lightCometClient.LightBlock(c.ctx, 0)
			if err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("get latest cl block failed")
				continue
			}

			if indexPoint.BlockHeight+10 > latestBlk.Height {
				continue
			}

			if err := c.index(indexPoint.BlockHeight+1, latestBlk.Height); err != nil {
				log.Error().Err(err).
					Str("indexer", c.Name()).
					Int64("from", indexPoint.BlockHeight).
					Int64("to", latestBlk.Height).
					Msg("index cl block failed")
			}
		}
	}
}

func (c *CLBlockIndexer) index(from, to int64) error {
	var blocks []*db.CLBlock

	for i := from; i <= to; i++ {
		blk, err := c.lightCometClient.LightBlock(c.ctx, i)
		if err != nil {
			return err
		}

		blocks = append(blocks, &db.CLBlock{
			Height:          blk.Height,
			Hash:            blk.Hash().String(),
			ProposerAddress: blk.ProposerAddress.String(),
			Time:            blk.Time,
		})

		if len(blocks) > 100 {
			if err := db.BatchCreateCLBlocks(c.dbOperator, c.Name(), blocks, i); err != nil {
				return err
			}

			blocks = make([]*db.CLBlock, 0)
		}
	}

	if len(blocks) > 0 {
		if err := db.BatchCreateCLBlocks(c.dbOperator, c.Name(), blocks, to); err != nil {
			return err
		}
	}

	return nil
}
