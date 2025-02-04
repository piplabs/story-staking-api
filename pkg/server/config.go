package server

import "fmt"

const (
	IndexModeReader = "reader"
	IndexModeWriter = "writer"
)

const (
	DatabaseEnginePostgres = "postgres"
)

const (
	CacheEngineRedis = "redis"
)

type Config struct {
	Blockchain BlockchainConfig `toml:"blockchain"`
	Server     ServerConfig     `toml:"server"`
	Database   DatabaseConfig   `toml:"database"`
	Cache      CacheConfig      `toml:"cache"`
}

type BlockchainConfig struct {
	ConsensusChainID    string `toml:"consensus_chain_id"`
	CometbftRPCEndpoint string `toml:"cometbft_rpc_endpoint"`
	StoryAPIEndpoint    string `toml:"story_api_endpoint"`
	GethRPCEndpoint     string `toml:"geth_rpc_endpoint"`
}

type ServerConfig struct {
	IndexMode   string `toml:"index_mode"`
	ServiceMode string `toml:"service_mode"`
	ServicePort string `toml:"service_port"`
}

type DatabaseConfig struct {
	Engine     string `toml:"engine"`
	ConfigFile string `toml:"config_file"`
}

type CacheConfig struct {
	Engine     string `toml:"engine"`
	ConfigFile string `toml:"config_file"`
}

func (c Config) Validate() error {
	switch c.Server.IndexMode {
	case IndexModeReader, IndexModeWriter:
		// Valid, do nothing.
	default:
		return fmt.Errorf("invalid index mode: %s", c.Server.IndexMode)
	}

	switch c.Database.Engine {
	case DatabaseEnginePostgres:
		// Valid, do nothing.
	default:
		return fmt.Errorf("invalid database engine: %s", c.Database.Engine)
	}

	switch c.Cache.Engine {
	case CacheEngineRedis:
		// Valid, do nothing.
	default:
		return fmt.Errorf("invalid cache engine: %s", c.Cache.Engine)
	}

	return nil
}
