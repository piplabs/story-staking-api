# API Documentation

## Indexed Data API

### 1. Network Status

[GET] `/api/network_status`

#### Response

```json
{
  "code": 200,
  "msg": {
    "status": "Normal",
    "consensus_block_height": 88,
    "execution_block_height": 88
  },
  "error": ""
}
```

### 2. Estimated APR

[GET] `/api/estimated_apr`

#### Response

```json
{
  "code": 200,
  "msg": {
    "apr": "1997%"
  },
  "error": ""
}
```

### 3. Operation History

[GET] `/api/operations/{evm_address}`

### Query Params

| Name                   | Type   | Example |
|------------------------|--------|---------|
| page                   | string | 1       |
| per_page               | string | 100     |

#### Response

```json
{
  "code": 200,
  "msg": {
    "operations": [
      {
        "tx_hash": "0x9f75c84b90e802c4218471ef4e1b68687b847394b1d6de5bbf4d29606d94d748",
        "block_height": 66,
        "event_type": "Unstake",
        "address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
        "src_validator_address": "",
        "dst_validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
        "dst_address": "",
        "status_ok": false,
        "error_code": "unspecified",
        "amount": "1024000000000"
      },
      {
        "tx_hash": "0xdf236f25a1544256cf829188b23ba62a938430aac408b4ace6fa97acde66f34d",
        "block_height": 64,
        "event_type": "Stake",
        "address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
        "src_validator_address": "",
        "dst_validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
        "dst_address": "",
        "status_ok": true,
        "error_code": "",
        "amount": "1024000000000"
      }
    ],
    "count": 2,
    "total": 2
  },
  "error": ""
}
```

### 4. Delegator Accumulated Rewards

[GET] `/api/rewards/{evm_address}`

#### Response

```json
{
  "code": 200,
  "msg": {
    "address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
    "amount": "450275716080",
    "last_update_height": 216
  },
  "error": ""
}
```

## Original Story API

### 1. Staking Pool

[GET] `/api/staking/pool`

#### Response

```json
{
  "code": 200,
  "msg": {
    "pool": {
      "not_bonded_tokens": "15308800000000",
      "bonded_tokens": "6144004000000"
    }
  },
  "error": ""
}
```

### 2. Validators Info

[GET] `/api/staking/validators`

### Query Params

| Name                   | Type   | Exampe                                                          |
|------------------------|--------|-----------------------------------------------------------------|
| status                 | string | BOND_STATUS_BONDED, BOND_STATUS_UNBONDING, BOND_STATUS_UNBONDED |
| pagination.key         | string |                                                                 |
| pagination.offset      | string | 100                                                             |
| pagination.limit       | string | 10                                                              |
| pagination.count_total | string | true                                                            |
| pagination.reverse     | string | false                                                           |

#### Response

```json
{
  "code": 200,
  "msg": {
    "validators": [
      {
        "operator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AvLo+lkg0UWozoI+pJzv1a7upt+HaMxZCdWgRxvZ8Cb1"
        },
        "jailed": false,
        "status": 3,
        "tokens": "6144001000000",
        "delegator_shares": "6144001000000.000000000000000000",
        "description": {
          "moniker": "0x0FC41199CE588948861A8DA86D725A5A073AE91A"
        },
        "uptime": "100%"
      },
      {
        "operator_address": "0x13665369a8ad5163f0c023839323b5d015925de1",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AlZ/RRQCXVnTPjZWqxXdaZ0X0ZvJvJRDHu1R4UpgwzXl"
        },
        "jailed": false,
        "status": 3,
        "tokens": "1000000",
        "delegator_shares": "1000000.000000000000000000",
        "description": {
          "moniker": "0x99C28AE30CBEFEFF75E91C66692FE0BD9279B861"
        },
        "uptime": "9.64%"
      },
      {
        "operator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AuVdgdSM9/7ezeZM/drHtdA0QVOboAaWEXQGoqDeRROt"
        },
        "jailed": true,
        "status": 2,
        "tokens": "972800000000",
        "delegator_shares": "1024000000000.000000000000000000",
        "description": {
          "moniker": "new_validator"
        },
        "uptime": "0%"
      },
      {
        "operator_address": "0x87f3cc50c84005f7130d37b849f6a71e05a8bf1f",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AzUWGMooEM92H8RCIOqXbjtbeur+2rOzgP9T/umnf1eA"
        },
        "jailed": false,
        "status": 3,
        "tokens": "1000000",
        "delegator_shares": "1000000.000000000000000000",
        "description": {
          "moniker": "0x9DFC26A7662106EEEC5E87B20CBB690CFCE73A05"
        },
        "uptime": "10.82%"
      },
      {
        "operator_address": "0xc47c28f925179089b6b7b1b336ac1f943b240066",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AqCVQjEtIzN9q5sMgSgl4dDD27vx6wa528lp9rjNKZE/"
        },
        "jailed": false,
        "status": 3,
        "tokens": "1000000",
        "delegator_shares": "1000000.000000000000000000",
        "description": {
          "moniker": "0x768A39103B552E7AE56635DD4E9B55922AAFC504"
        },
        "uptime": "9.17%"
      }
    ],
    "pagination": {
      "next_key": "",
      "total": "5"
    }
  },
  "error": ""
}
```

### 3. Validator Info

[GET] `/api/staking/validators/{validator_address}`

#### Response

```json
{
  "code": 200,
  "msg": {
    "operator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
    "consensus_pubkey": {
      "type": "tendermint/PubKeySecp256k1",
      "value": "AvLo+lkg0UWozoI+pJzv1a7upt+HaMxZCdWgRxvZ8Cb1"
    },
    "jailed": false,
    "status": 3,
    "tokens": "6144001000000",
    "delegator_shares": "6144001000000.000000000000000000",
    "description": {
      "moniker": "0x0FC41199CE588948861A8DA86D725A5A073AE91A"
    },
    "uptime": "100%"
  },
  "error": ""
}
```

### 4. Delegations of a Validator

[GET] `/api/staking/validators/{validator_address}/delegations`

### Query Params

| Name                   | Type   | Example |
|------------------------|--------|---------|
| pagination.key         | string |         |
| pagination.offset      | string | 100     |
| pagination.limit       | string | 10      |
| pagination.count_total | string | true    |
| pagination.reverse     | string | false   |

#### Response

```json
{
  "code": 200,
  "msg": {
    "delegation_responses": [
      {
        "delegation": {
          "delegator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
          "validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
          "shares": "1000000.000000000000000000",
          "rewards_shares": "500000.000000000000000000"
        },
        "balance": {
          "denom": "stake",
          "amount": "1000000"
        }
      },
      {
        "delegation": {
          "delegator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
          "validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
          "shares": "6144000000000.000000000000000000",
          "rewards_shares": "3072000000000.000000000000000000"
        },
        "balance": {
          "denom": "stake",
          "amount": "6144000000000"
        }
      }
    ],
    "pagination": {
      "next_key": "",
      "total": "2"
    }
  },
  "error": ""
}
```

### 5. Delegation of a Validator

[GET] `/api/staking/validators/{validator_address}/delegations/{delegator_address}`

#### Response

```json
{
  "code": 200,
  "msg": {
    "delegation_response": {
      "delegation": {
        "delegator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
        "validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
        "shares": "6144000000000.000000000000000000",
        "rewards_shares": "3072000000000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "6144000000000"
      }
    }
  },
  "error": ""
}
```

### 6. Period Delegations of a Validator

[GET] `/api/staking/validators/{validator_address}/delegators/{delegator_address}/period_delegations`

### Query Params

| Name                   | Type   | Example |
|------------------------|--------|---------|
| pagination.key         | string |         |
| pagination.offset      | string | 100     |
| pagination.limit       | string | 10      |
| pagination.count_total | string | true    |
| pagination.reverse     | string | false   |

#### Response

```json
{
  "code": 200,
  "msg": {
    "period_delegation_responses": [
      {
        "period_delegation": {
          "delegator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
          "validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
          "period_delegation_id": "0",
          "shares": "6144000000000.000000000000000000",
          "rewards_shares": "3072000000000.000000000000000000",
          "end_time": "2025-01-31T23:16:16.421292251Z"
        },
        "balance": {
          "denom": "stake",
          "amount": "6144000000000"
        }
      }
    ],
    "pagination": {
      "next_key": "",
      "total": "1"
    }
  },
  "error": ""
}
```

### 7. Period Delegation of a Validator

[GET] `/api/staking/validators/{validator_address}/delegators/{delegator_address}/period_delegations/{period_delegation_id}`

#### Response

```json
{
  "code": 200,
  "msg": {
    "period_delegation_response": {
      "period_delegation": {
        "delegator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
        "validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
        "period_delegation_id": "0",
        "shares": "6144000000000.000000000000000000",
        "rewards_shares": "3072000000000.000000000000000000",
        "end_time": "2025-01-31T23:16:16.421292251Z"
      },
      "balance": {
        "denom": "stake",
        "amount": "6144000000000"
      }
    }
  },
  "error": ""
}
```

### 8. Delegations of a Delegator

[GET] `/api/staking/delegations/{delegator_address}`

### Query Params

| Name                   | Type   | Example |
|------------------------|--------|---------|
| pagination.key         | string |         |
| pagination.offset      | string | 100     |
| pagination.limit       | string | 10      |
| pagination.count_total | string | true    |
| pagination.reverse     | string | false   |

#### Response

```json
{
  "code": 200,
  "msg": {
    "delegation_responses": [
      {
        "delegation": {
          "delegator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
          "validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
          "shares": "6144000000000.000000000000000000",
          "rewards_shares": "3072000000000.000000000000000000"
        },
        "balance": {
          "denom": "stake",
          "amount": "6144000000000"
        }
      },
      {
        "delegation": {
          "delegator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
          "validator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
          "shares": "1024000000000.000000000000000000",
          "rewards_shares": "1024000000000.000000000000000000"
        },
        "balance": {
          "denom": "stake",
          "amount": "972800000000"
        }
      }
    ],
    "pagination": {
      "next_key": "",
      "total": "2"
    }
  },
  "error": ""
}
```

### 9. Unbonding Delegations of a Delegator

[GET] `/api/staking/delegators/{delegator_address}/unbonding_delegations`

### Query Params

| Name                   | Type   | Example |
|------------------------|--------|---------|
| pagination.key         | string |         |
| pagination.offset      | string | 100     |
| pagination.limit       | string | 10      |
| pagination.count_total | string | true    |
| pagination.reverse     | string | false   |

#### Response

```json
{
  "code": 200,
  "msg": {
    "unbonding_responses": [
      {
        "delegator_address": "0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73",
        "validator_address": "0x00a842dbd3d11176b4868dd753a552b8919d5a63",
        "entries": [
          {
            "creation_height": "15",
            "completion_time": "2025-02-01T02:03:07.77114334Z",
            "initial_balance": "1024000000000",
            "balance": "1024000000000",
            "unbonding_id": "1"
          },
          {
            "creation_height": "16",
            "completion_time": "2025-02-01T02:03:10.403828008Z",
            "initial_balance": "1024000000000",
            "balance": "1024000000000",
            "unbonding_id": "2"
          }
        ]
      }
    ],
    "pagination": {
      "next_key": "",
      "total": "1"
    }
  },
  "error": ""
}
```
