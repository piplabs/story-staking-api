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
	if errors.Is(err, redis.Nil) {
		return nil, false
	} else if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to get data from cache")
		return nil, false
	}

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

	if err := cache.SetRedisData(ctx, rdb, key, string(jsonData)); err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to cache data")
		return false
	}

	return true
}
