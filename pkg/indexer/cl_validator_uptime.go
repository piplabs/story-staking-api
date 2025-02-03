package indexer

import (
	"context"
	"strings"
	"time"

	lightprovider "github.com/cometbft/cometbft/light/provider"
	lighthttp "github.com/cometbft/cometbft/light/provider/http"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/types"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/cache"
	"github.com/piplabs/story-staking-api/db"
	"github.com/piplabs/story-staking-api/pkg/util"
)

var _ Indexer = (*CLValidatorUptimeIndexer)(nil)

type CLValidatorUptimeIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	cometClient      *comethttp.HTTP
	lightCometClient lightprovider.Provider
}

func NewCLValidatorUptimeIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, chainID, rpcEndpoint string) (*CLValidatorUptimeIndexer, error) {
	cometClient, err := comethttp.New(rpcEndpoint, "")
	if err != nil {
		return nil, err
	}

	lightCometClient, err := lighthttp.New(chainID, rpcEndpoint)
	if err != nil {
		return nil, err
	}

	return &CLValidatorUptimeIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		cometClient:      cometClient,
		lightCometClient: lightCometClient,
	}, nil
}

func (c *CLValidatorUptimeIndexer) Name() string {
	return "cl_validator_uptime"
}

func (c *CLValidatorUptimeIndexer) Run() {
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

func (c *CLValidatorUptimeIndexer) index(from, to int64) error {
	validatorUptime := make(map[string]*db.CLValidatorUptime)

	for i := from; i <= to; i++ {
		if err := c.fetchActiveValidators(c.ctx, i, validatorUptime); err != nil {
			return err
		}

		if err := c.fetchVotes(c.ctx, i, validatorUptime); err != nil {
			return err
		}
	}

	clUptimes := make([]*db.CLValidatorUptime, 0, len(validatorUptime))
	for _, uptime := range validatorUptime {
		clUptimes = append(clUptimes, uptime)
	}

	c.invalidateCache()

	if err := db.BatchUpsertCLValidatorUptime(c.dbOperator, c.Name(), clUptimes, to); err != nil {
		return err
	}

	return nil
}

func (c *CLValidatorUptimeIndexer) fetchActiveValidators(ctx context.Context, height int64, validatorUptime map[string]*db.CLValidatorUptime) error {
	page, perPage := 1, 100
	for {
		validatorsRes, err := c.cometClient.Validators(ctx, &height, &page, &perPage)
		if err != nil {
			return err
		}

		for _, validator := range validatorsRes.Validators {
			cometAddr := validator.Address.String()

			evmAddr, err := util.CmpPubKeyToEVMAddress(validator.PubKey.Bytes())
			if err != nil {
				return err
			}

			uptime, ok := validatorUptime[cometAddr]
			if !ok || uptime.ActiveTo != height-1 {
				validatorUptime[cometAddr] = &db.CLValidatorUptime{
					EVMAddress: strings.ToLower(evmAddr.String()),
					ActiveFrom: height,
					ActiveTo:   height,
					VoteCount:  0,
				}
			} else {
				uptime.ActiveTo = height
			}
		}

		if page*perPage >= validatorsRes.Total {
			break
		}
		page++
	}

	return nil
}

func (c *CLValidatorUptimeIndexer) fetchVotes(ctx context.Context, height int64, validatorUptime map[string]*db.CLValidatorUptime) error {
	commitRes, err := c.cometClient.Commit(ctx, &height)
	if err != nil {
		return err
	}

	for _, sig := range commitRes.Commit.Signatures {
		if sig.BlockIDFlag != types.BlockIDFlagCommit {
			continue
		}

		if uptime, ok := validatorUptime[sig.ValidatorAddress.String()]; ok {
			uptime.VoteCount++
		}
	}

	return nil
}

func (c *CLValidatorUptimeIndexer) invalidateCache() {
	_ = cache.InvalidateRedisDataByPrefix(c.ctx, c.cacheOperator, cache.ValidatorsKeyPrefix)
}
