package indexer

import (
	"context"
	"time"

	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

var _ Indexer = (*CLBlockIndexer)(nil)

type CLBlockIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	cometClient *comethttp.HTTP
}

func NewCLBlockIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, rpcEndpoint string) (*CLBlockIndexer, error) {
	cometClient, err := comethttp.New(rpcEndpoint, "")
	if err != nil {
		return nil, err
	}

	return &CLBlockIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		cometClient: cometClient,
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

			latestBlk, err := c.cometClient.Block(c.ctx, nil)
			if err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("get latest cl block failed")
				continue
			}

			if indexPoint.BlockHeight+10 > latestBlk.Block.Height {
				continue
			}

			if err := c.index(indexPoint.BlockHeight+1, latestBlk.Block.Height); err != nil {
				log.Error().Err(err).
					Str("indexer", c.Name()).
					Int64("from", indexPoint.BlockHeight).
					Int64("to", latestBlk.Block.Height).
					Msg("index cl block failed")
			}
		}
	}
}

func (c *CLBlockIndexer) index(from, to int64) error {
	var blocks []*db.CLBlock

	for i := from; i <= to; i++ {
		blk, err := c.cometClient.Block(c.ctx, &i)
		if err != nil {
			return err
		}

		blocks = append(blocks, &db.CLBlock{
			Height:          blk.Block.Height,
			Hash:            blk.Block.Hash().String(),
			ProposerAddress: blk.Block.ProposerAddress.String(),
			Time:            blk.Block.Time,
		})

		if len(blocks) > 100 {
			if err := db.BatchCreateCLBlocks(c.dbOperator, c.Name(), blocks, i); err != nil {
				return err
			}

			blocks = make([]*db.CLBlock, 0)
		}
	}

	// Handle remaining entries, even if there are no entries, we also need to update the index point.
	if err := db.BatchCreateCLBlocks(c.dbOperator, c.Name(), blocks, to); err != nil {
		return err
	}

	return nil
}
