# Story Staking API

**Story Staking API** is a sophisticated data aggregation and processing tool that interfaces with various layers of nodes to compile and refine collected data, ultimately outputting high-dimensional scenario indicators.

## Prerequisites

- [Go1.23+](https://golang.org/dl/)

## Features

The primary function of the Story Staking API is to provide interfaces through which it aggregates and processes data. These interfaces include various modules, each dedicated to different aspects of network data. As of now, the following module has been fully implemented:

### Staking API
- **Network Status**: Provides the status of the network.
- **Network Estimated APR**: Provides an estimate of the Annual Percentage Rate (APR) for the network.
- **Operation History**: Provides a list of stakingoperations for a given address.
- **Delegator Accumulated $IP Rewards**: Summarizes the total rewards earned by a delegator.
- **Validator Uptime**: Tracks the uptime of a validator.

## Build and Run

### Build
```bash
$ make build
$ ./story-staking-api --help
usage: story-staking-api [<flags>]

Flags:
  --[no-]help             Show context-sensitive help (also try --help-long and --help-man).
  --home="."              Home directory
  --config="config.toml"  Config file path
```

### Configuration

Below is an example of a configuration file (`config.toml`). Please replace with your actual parameters:

```toml
[blockchain]
consensus_chain_id = "story-localnet"
cometbft_rpc_endpoint = "http://localhost:26657"
story_api_endpoint = "http://localhost:1317"
geth_rpc_endpoint = "http://localhost:8545"

[server]
# Index mode options: reader | writer
index_mode = "writer"
# Service mode options: debug | release
service_mode = "debug"
service_port = ":8080"

[database]
# Database engine: postgres | mysql
engine = "postgres"
config_file = "config/postgres.yaml"

[cache]
# Cache engine: redis | freecache
# * redis (github.com/redis/go-redis/v9)
# * freecache (github.com/coocood/freecache)
engine = "redis"
config_file = "config/redis.yaml"
```

Note that the `config_file` path is relative to the home directory specified by the `--home` flag when running the binary.

### API Documentation

Please refer to [API Documentation](./pkg/server/README.md).
