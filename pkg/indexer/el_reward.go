package indexer

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-indexer/cache"
	"github.com/piplabs/story-indexer/db"
)

var _ Indexer = (*ELRewardIndexer)(nil)

type ELRewardIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	ethClient *ethclient.Client
}

func NewELRewardIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, rpcEndpoint string) (*ELRewardIndexer, error) {
	ethClient, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return nil, err
	}

	return &ELRewardIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		ethClient: ethClient,
	}, nil
}

func (e *ELRewardIndexer) Name() string {
	return "el_reward"
}

func (e *ELRewardIndexer) Run() {
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
					Int64("from", indexPoint.BlockHeight).
					Uint64("to", latestBlkNum).
					Msg("index el reward failed")
			}
		}
	}
}

func (e *ELRewardIndexer) index(from, to int64) error {
	elRewardsMap := make(map[string]*db.ELReward)

	for i := from; i <= to; i++ {
		blk, err := e.ethClient.BlockByNumber(e.ctx, big.NewInt(i))
		if err != nil {
			return err
		}

		for _, w := range blk.Withdrawals() {
			address := strings.ToLower(w.Address.String())

			if _, ok := elRewardsMap[address]; ok {
				elRewardsMap[address].Amount += int64(w.Amount)
			} else {
				elRewardsMap[address] = &db.ELReward{
					Address:          address,
					Amount:           int64(w.Amount),
					LastUpdateHeight: i,
				}
			}
		}

		if len(elRewardsMap) > 100 {
			elRewards := make([]*db.ELReward, 0, len(elRewardsMap))
			for _, v := range elRewardsMap {
				elRewards = append(elRewards, v)
			}

			e.invalidateCache(elRewards)

			if err := db.BatchUpsertELRewards(e.dbOperator, e.Name(), elRewards, i); err != nil {
				return err
			}

			elRewardsMap = make(map[string]*db.ELReward)
		}
	}

	// Handle remaining entries, even if there are no entries, we also need to update the index point.
	elRewards := make([]*db.ELReward, 0, len(elRewardsMap))
	for _, v := range elRewardsMap {
		elRewards = append(elRewards, v)
	}

	e.invalidateCache(elRewards)

	if err := db.BatchUpsertELRewards(e.dbOperator, e.Name(), elRewards, to); err != nil {
		return err
	}

	return nil
}

func (e *ELRewardIndexer) invalidateCache(elRewards []*db.ELReward) {
	for _, v := range elRewards {
		_ = cache.InvalidateRedisData(e.ctx, e.cacheOperator, cache.RewardsKey(v.Address))
	}
}
