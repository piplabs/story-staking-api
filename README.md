# Story Staking API

**Story Staking API** is a sophisticated data aggregation and processing tool that interfaces with various layers of nodes to compile and refine collected data, ultimately outputting high-dimensional scenario indicators.

## Prerequisites

- [Go1.23.4+](https://golang.org/dl/)

## Features

The primary function of the Story Staking API is to provide interfaces through which it aggregates and processes data. These interfaces include various modules, each dedicated to different aspects of network data. As of now, the following module has been fully implemented:

### Staking API

- **Network Status**: Provides the status of the network.
- **Network Estimated APR**: Provides an estimate of the Annual Percentage Rate (APR) for the network.
- **Operation History**: Provides a list of stakingoperations for a given address.
- **Delegator Accumulated $IP Rewards**: Summarizes the total rewards earned by a delegator.
- **Validator Uptime**: Tracks the uptime of a validator.

## Build and Run

### Prerequisites

1. **Database (Postgres or MySQL)**
   You can use your own database or run the following command to quickly set up a local Postgres container:

   ```bash
   # Spin up a local Postgres container
   make local-postgres
   ```

   > If you prefer MySQL, ensure you have a local or remote instance running.

2. **Redis Cache**
   You can use your own Redis server or run the following command to quickly set up a local Redis container:
   ```bash
   # Spin up a local Redis container
   make local-redis
   ```

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

#### DB Credentials

The `config_file` fields in the `[database]` and `[cache]` sections are optional. You can also specify credentials via environment variables. For example:

```bash
# Redis environment variables
REDIS_PASSWORD_MODE="plain"
REDIS_DB="0"
REDIS_ADDR=""
REDIS_PASSWORD=""

# Postgres environment variables
POSTGRES_DB_USERNAME=""
POSTGRES_DB_PASSWORD=""
POSTGRES_DB_PASSWORD_MODE="plain"
POSTGRES_DB_HOST=""
POSTGRES_DB_PORT=""
POSTGRES_DB_NAME=""
POSTGRES_DB_MAX_CONNS="100"
POSTGRES_DB_MIN_CONNS="0"
POSTGRES_DB_MAX_CONN_LIFETIME="12h"
POSTGRES_DB_MAX_CONN_IDLE_TIME="5m"
```

> **Note:** Adjust each value as needed for your setup.

#### Run in Cluster Mode

**Staking API** supports running in a cluster mode, with a designated **writer** node and one or more **reader** nodes. Below are example configuration files:

**Writer Configuration**

Create a `writer-config.toml` with the following content:

```toml
[blockchain]
consensus_chain_id = ""
cometbft_rpc_endpoint = ""
story_api_endpoint = ""
geth_rpc_endpoint = ""

[server]
index_mode = "writer"
service_mode = "debug"
service_port = ":8080"

[database]
engine = "postgres"
config_file = ""

[cache]
engine = "redis"
config_file = ""
```

**Reader Configuration**

Create one `reader-config.toml` files to scale read operations as needed:

```toml
[blockchain]
consensus_chain_id = ""
cometbft_rpc_endpoint = ""
story_api_endpoint = ""
geth_rpc_endpoint = ""

[server]
index_mode = "reader"
service_mode = "debug"
service_port = ":8080"

[database]
engine = "postgres"
config_file = ""

[cache]
engine = "redis"
config_file = ""
```

Note that the `config_file` path is relative to the home directory specified by the `--home` flag when running the binary.

### API Documentation

Please refer to [API Documentation](./pkg/server/README.md).
