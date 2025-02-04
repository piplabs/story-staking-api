package db

import (
	"context"
	"fmt"
	"hash/crc32"
	"os"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/kelseyhightower/envconfig"
	yaml "gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DefaultPostgresConfigPrefix = "postgres"

	DBPasswordModePlain = "plain"
	DBPasswordModeGCP   = "gcp-secret-manager"
)

type PostgresConfig struct {
	DBUsername     string `yaml:"db-username" envconfig:"DB_USERNAME"`
	DBPasswordMode string `yaml:"db-password-mode" envconfig:"DB_PASSWORD_MODE"`
	DBPassword     string `yaml:"db-password" envconfig:"DB_PASSWORD"`
	DBHost         string `yaml:"db-host" envconfig:"DB_HOST"`
	DBPort         int    `yaml:"db-port" envconfig:"DB_PORT"`
	DBName         string `yaml:"db-name" envconfig:"DB_NAME"`

	DBMaxConns        int           `yaml:"db-max-conns" envconfig:"DB_MAX_CONNS"`
	DBMinConns        int           `yaml:"db-min-conns" envconfig:"DB_MIN_CONNS"`
	DBMaxConnLifetime time.Duration `yaml:"db-max-conn-lifetime" envconfig:"DB_MAX_CONN_LIFETIME"`
	DBMaxConnIdleTime time.Duration `yaml:"db-max-conn-idle-time" envconfig:"DB_MAX_CONN_IDLE_TIME"`
}

func NewPostgresClient(ctx context.Context, configFile string) (*gorm.DB, error) {
	var config PostgresConfig

	if configFile != "" {
		f, err := os.Open(configFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		if err := yaml.NewDecoder(f).Decode(&config); err != nil {
			return nil, err
		}
	}

	// Load environment variables
	if err := envconfig.Process(DefaultPostgresConfigPrefix, &config); err != nil {
		return nil, err
	}

	var dbPassword string
	switch config.DBPasswordMode {
	case DBPasswordModePlain:
		dbPassword = config.DBPassword
	case DBPasswordModeGCP:
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create secretmanager client: %w", err)
		}
		defer client.Close()

		result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: config.DBPassword})
		if err != nil {
			return nil, fmt.Errorf("failed to access secret version: %w", err)
		}

		crc32c := crc32.MakeTable(crc32.Castagnoli)
		checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
		if checksum != *result.Payload.DataCrc32C {
			return nil, fmt.Errorf("data corruption detected")
		}
		dbPassword = string(result.Payload.Data)
	default:
		return nil, fmt.Errorf("invalid db password mode: %s", config.DBPasswordMode)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		config.DBHost,
		config.DBUsername,
		dbPassword,
		config.DBName,
		config.DBPort,
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(config.DBMaxConns)
	sqlDB.SetMaxIdleConns(config.DBMinConns)
	sqlDB.SetConnMaxLifetime(config.DBMaxConnLifetime)
	sqlDB.SetConnMaxIdleTime(config.DBMaxConnIdleTime)

	return db, nil
}
