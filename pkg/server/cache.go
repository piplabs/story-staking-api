package server

import (
	"context"
	"encoding/json"
	"errors"

	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/piplabs/story-staking-api/cache"
)

func GetCachedData[T any](ctx context.Context, rdb *redis.Client, key string) (*T, bool) {
	cacheData, err := cache.GetRedisData(ctx, rdb, key)
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error().Err(err).Str("key", key).Msg("failed to get data from cache")
		return nil, false
	}

	log.Info().Str("key", key).Str("cache_data", cacheData).Msg("get cached data")

	var result T
	if err := json.Unmarshal([]byte(cacheData), &result); err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to unmarshal cached data")
		return nil, false
	}

	return &result, true
}

func SetCachedData[T any](ctx context.Context, rdb *redis.Client, key string, data T) bool {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to marshal data")
		return false
	}

	log.Info().Str("key", key).Str("cache_data", string(jsonData)).Msg("set cached data")

	if err := cache.SetRedisData(ctx, rdb, key, string(jsonData)); err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to cache data")
		return false
	}

	return true
}
