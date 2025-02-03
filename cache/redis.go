package cache

import (
	"context"
	"fmt"
	"hash/crc32"
	"os"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/kelseyhightower/envconfig"
	redis "github.com/redis/go-redis/v9"
	yaml "gopkg.in/yaml.v2"
)

const (
	DefaultCacheTTL = time.Hour
)

const (
	DefaultRedisConfigPrefix = "redis"

	RedisPasswordModePlain = "plain"
	RedisPasswordModeGCP   = "gcp-secret-manager"
)

type RedisConfig struct {
	Addr         string `yaml:"addr" envconfig:"ADDR"`
	PasswordMode string `yaml:"password-mode" envconfig:"PASSWORD_MODE"`
	Password     string `yaml:"password" envconfig:"PASSWORD"`
	DB           int    `yaml:"db" envconfig:"DB"`
}

func NewRedisClient(ctx context.Context, configFile string) (*redis.Client, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config RedisConfig
	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	// Load environment variables
	if err := envconfig.Process(DefaultRedisConfigPrefix, &config); err != nil {
		return nil, err
	}

	var redisPassword string
	switch config.PasswordMode {
	case RedisPasswordModePlain:
		redisPassword = config.Password
	case RedisPasswordModeGCP:
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create secretmanager client: %w", err)
		}
		defer client.Close()

		result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: config.Password})
		if err != nil {
			return nil, fmt.Errorf("failed to access secret version: %w", err)
		}

		crc32c := crc32.MakeTable(crc32.Castagnoli)
		checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
		if checksum != *result.Payload.DataCrc32C {
			return nil, fmt.Errorf("data corruption detected")
		}
		redisPassword = string(result.Payload.Data)
	default:
		return nil, fmt.Errorf("invalid redis password mode: %s", config.PasswordMode)
	}

	return redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: redisPassword,
		DB:       config.DB,
	}), nil
}

func GetRedisData(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	return rdb.Get(ctx, key).Result()
}

func SetRedisData(ctx context.Context, rdb *redis.Client, key string, data string) error {
	return rdb.Set(ctx, key, data, DefaultCacheTTL).Err()
}

func InvalidateRedisData(ctx context.Context, rdb *redis.Client, key string) error {
	return rdb.Del(ctx, key).Err()
}

func InvalidateRedisDataByPrefix(ctx context.Context, rdb *redis.Client, prefix string) error {
	var (
		cursor    uint64
		batchSize int64 = 100
	)

	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, prefix+"*", batchSize).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			// Delete keys in batch
			if err := rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}

	return nil
}
