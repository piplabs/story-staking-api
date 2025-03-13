package indexer

import (
	"context"
	"fmt"
	"strings"
	"time"

	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/types"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/cache"
	"github.com/piplabs/story-staking-api/db"
	"github.com/piplabs/story-staking-api/pkg/util"
)

var _ Indexer = (*CLValidatorVoteIndexer)(nil)

type CLValidatorVoteIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	cometClient *comethttp.HTTP
}

func NewCLValidatorVoteIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, rpcEndpoint string) (*CLValidatorVoteIndexer, error) {
	cometClient, err := comethttp.New(rpcEndpoint, "")
	if err != nil {
		return nil, err
	}

	return &CLValidatorVoteIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		cometClient: cometClient,
	}, nil
}

func (c *CLValidatorVoteIndexer) Name() string {
	return "cl_validator_vote"
}

func (c *CLValidatorVoteIndexer) Run() {
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
					Msg("index cl validator votes failed")
			}
		}
	}
}

func (c *CLValidatorVoteIndexer) index(from, to int64) error {
	start := from

	for start <= to {
		end := min(start+100, to)

		validatorVotes := make([]*db.CLValidatorVote, 0)
		for i := start; i <= end; i++ {
			valVotes, err := c.fetchValidatorVotes(c.ctx, i)
			if err != nil {
				return err
			}

			validatorVotes = append(validatorVotes, valVotes...)
		}

		if err := db.BatchUpdateCLValidatorVotes(c.dbOperator, c.Name(), validatorVotes, end); err != nil {
			return err
		}

		start = end + 1
	}

	c.invalidateCache()

	return nil
}

func (c *CLValidatorVoteIndexer) fetchValidatorVotes(ctx context.Context, height int64) ([]*db.CLValidatorVote, error) {
	validatorVotes := make([]*db.CLValidatorVote, 0)

	cometAddrToEVMAddr, err := c.fetchActiveValidators(ctx, height)
	if err != nil {
		return nil, err
	}

	commitRes, err := c.cometClient.Commit(ctx, &height)
	if err != nil {
		return nil, err
	}

	for _, sig := range commitRes.Commit.Signatures {
		if !(sig.BlockIDFlag == types.BlockIDFlagCommit || sig.BlockIDFlag == types.BlockIDFlagNil) {
			continue
		}

		cometAddr := sig.ValidatorAddress.String()

		evmAddr, ok := cometAddrToEVMAddr[cometAddr]
		if !ok {
			return nil, fmt.Errorf("validator %s in block %d not found in active validators", cometAddr, height)
		}

		validatorVotes = append(validatorVotes, &db.CLValidatorVote{
			Validator:   strings.ToLower(evmAddr),
			BlockHeight: height,
		})
	}

	return validatorVotes, nil
}

func (c *CLValidatorVoteIndexer) fetchActiveValidators(ctx context.Context, height int64) (map[string]string, error) {
	cometAddrToEVMAddr := make(map[string]string)

	page, perPage := 1, 100
	for {
		validatorsRes, err := c.cometClient.Validators(ctx, &height, &page, &perPage)
		if err != nil {
			return nil, err
		}

		for _, validator := range validatorsRes.Validators {
			cometAddr := validator.Address.String()

			evmAddr, err := util.CmpPubKeyToEVMAddress(validator.PubKey.Bytes())
			if err != nil {
				return nil, err
			}

			cometAddrToEVMAddr[cometAddr] = evmAddr.String()
		}

		if page*perPage >= validatorsRes.Total {
			break
		}
		page++
	}

	return cometAddrToEVMAddr, nil
}

func (c *CLValidatorVoteIndexer) invalidateCache() {
	_ = cache.InvalidateRedisDataByPrefix(c.ctx, c.cacheOperator, cache.ValidatorsKeyPrefix)
}
