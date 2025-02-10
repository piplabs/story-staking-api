# API Documentation
- [Indexed Data API](#indexed-data-api)
  - [1. Network Status](#1-network-status)
  - [2. Estimated APR](#2-estimated-apr)
  - [3. Operation History](#3-operation-history)
  - [4. Delegator Accumulated Rewards](#4-delegator-accumulated-rewards)
- [Native Story API](#native-story-api)
  - [1. Staking Pool](#1-staking-pool)
  - [2. Validators Info](#2-validators-info)
  - [3. Validator Info](#3-validator-info)
  - [4. Delegations of a Validator](#4-delegations-of-a-validator)
  - [5. Delegation of a Validator](#5-delegation-of-a-validator)
  - [6. Period Delegations of a Validator](#6-period-delegations-of-a-validator)
  - [7. Period Delegation of a Validator](#7-period-delegation-of-a-validator)
  - [8. Delegations of a Delegator](#8-delegations-of-a-delegator)
  - [9. Unbonding Delegations of a Delegator](#9-unbonding-delegations-of-a-delegator)

## Indexed Data API

### 1. Network Status

[GET] `/api/network_status`

#### Response

- status: The status of the network.
  - `Normal`: The network is running normally.
  - `Degraded`: The network is congested but still operational.
  - `Down`: The network is experiencing critical issues and is no longer operational.
- consensus_block_height: The latest block height of the consensus layer.
- execution_block_height: The latest block height of the execution layer.

```json
{
  "code": 200,
  "msg": {
    "status": "Normal",
    "consensus_block_height": 89,
    "execution_block_height": 88
  },
  "error": ""
}
```

### 2. Estimated APR

[GET] `/api/estimated_apr`

#### Response

- apr: The estimated APR of the network.

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

#### Path Params

| Name           | Type   | Example                                    | Required |
|----------------|--------|--------------------------------------------|----------|
| evm_address    | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Query Params

| Name                   | Type   | Example | Required |
|------------------------|--------|---------|----------|
| page                   | string | 1       | No       |
| per_page               | string | 100     | No       |

#### Response

- operations: The list of operations.
  - status_ok: Whether the operation is successful.
  - error_code: The error code of the operation, non-empty for failed operations.
  - amount: The amount of the operation in `gwei` .
  - tx_hash: The hash of the transaction.
  - block_height: The block height of the transaction.
  - event_type: The type of the event.
  - address: The address that performs the operation.
  - src_validator_address: The source validator address, non-empty for `Redelegate` and `RedelegateOnBehalf` events.
  - dst_validator_address: The destination validator address, non-empty for `Stake`, `StakeOnBehalf`, `Redelegate`, `RedelegateOnBehalf`, `Unstake`, `UnstakeOnBehalf`, `CreateValidator`, `Unjail`, `UnjailOnBehalf` and `UpdateValidatorCommission` events.
  - dst_address: The destination address, non-empty for `SetOperator`, `SetWithdrawalAddress` and `SetRewardAddress` events.
- count: The number of operations in the current page.
- total: The total number of operations.

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

#### Path Params

| Name           | Type   | Example                                    | Required |
|----------------|--------|--------------------------------------------|----------|
| evm_address    | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Response

- address: The address of the delegator.
- amount: The accumulated rewards of the delegator in `gwei`.
- last_update_height: Last updated block height of the rewards.

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

## Native Story API

### 1. Staking Pool

[GET] `/api/staking/pool`

#### Response

- pool: The staking pool info.
  - bonded_tokens: The total number of bonded tokens in `gwei`.
  - not_bonded_tokens: The total number of not bonded tokens in `gwei`.

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

#### Query Params

| Name                   | Type   | Example                                                               |
|------------------------|--------|-----------------------------------------------------------------------|
| status                 | string | `BOND_STATUS_UNBONDED`, `BOND_STATUS_UNBONDING`, `BOND_STATUS_BONDED` |
| pagination.key         | string |                                                                       |
| pagination.offset      | string | 100                                                                   |
| pagination.limit       | string | 10                                                                    |
| pagination.count_total | string | true                                                                  |
| pagination.reverse     | string | false                                                                 |

#### Response

- validators: The list of validators.
  - operator_address: The evm address of the validator.
  - consensus_pubkey: The base64 encoded compressed public key of the validator.
  - jailed: Whether the validator is jailed.
  - status: The status of the validator.
    - 1: `BOND_STATUS_UNBONDED`
    - 2: `BOND_STATUS_UNBONDING`
    - 3: `BOND_STATUS_BONDED`
  - tokens: The total staked tokens on the validator in `gwei`.
  - delegator_shares: The total delegator shares on the validator.
  - description: The description of the validator.
    - moniker: The moniker of the validator.
  - commission: The commission of the validator.
  - support_token_type: The support token type of the validator.
    - 0: `LOCKED`
    - 1: `UNLOCKED`
  - uptime: The uptime of the validator, empty if the validator has never been bonded.
  - apr: The apr of the validator, affected by the network apr and the validator's commission rate.
- pagination: The pagination info.
  - next_key: The key to query the next page.
  - total: The total number of validators.

```json
{
  "code": 200,
  "msg": {
    "validators": [
      {
        "operator_address": "0xc5c0beeac8b37ed52f6a675ee2154d926a88e3ec",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AqBVHHkyOfiie29Wrez6hMvC644kbZfPgXA1jFEs7Uwq"
        },
        "jailed": false,
        "status": 3,
        "tokens": "10039668001000000",
        "delegator_shares": "10039668001000000.000000000000000000",
        "description": {
          "moniker": "0x0FC41199CE588948861A8DA86D725A5A073AE91A"
        },
        "commission": {
          "commission_rates": {
            "rate": "0.070000000000000000",
            "max_rate": "0.100000000000000000",
            "max_change_rate": "0.010000000000000000"
          },
          "update_time": "2025-01-19T15:00:00Z"
        },
        "support_token_type": 0,
        "uptime": "98.84%",
        "apr": "18.43%"
      },
      {
        "operator_address": "0xcd29b70ff04c0aa386f7b3453df0e5ed3d4f67bb",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "A9mdRUGE+sv2oD6jfrNvalDGmELqOtQgOKjVU3vRWyWU"
        },
        "jailed": false,
        "status": 3,
        "tokens": "20317549001000000",
        "delegator_shares": "20317549001000000.000000000000000000",
        "description": {
          "moniker": "0x768A39103B552E7AE56635DD4E9B55922AAFC504"
        },
        "commission": {
          "commission_rates": {
            "rate": "0.070000000000000000",
            "max_rate": "0.100000000000000000",
            "max_change_rate": "0.010000000000000000"
          },
          "update_time": "2025-01-19T15:00:00Z"
        },
        "support_token_type": 1,
        "uptime": "99.8%",
        "apr": "36.86%"
      },
      {
        "operator_address": "0xcd5faabca5bea3c5fc5e2371c7b397604720c2c2",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "A/9SMxZTnh3Rq96Eygg9MfB6g82euMXhjT5nMWrhLlyf"
        },
        "jailed": false,
        "status": 3,
        "tokens": "10035524001000000",
        "delegator_shares": "10035524001000000.000000000000000000",
        "description": {
          "moniker": "0x99C28AE30CBEFEFF75E91C66692FE0BD9279B861"
        },
        "commission": {
          "commission_rates": {
            "rate": "0.070000000000000000",
            "max_rate": "0.100000000000000000",
            "max_change_rate": "0.010000000000000000"
          },
          "update_time": "2025-01-19T15:00:00Z"
        },
        "support_token_type": 0,
        "uptime": "98.64%",
        "apr": "18.43%"
      },
      {
        "operator_address": "0xdb8e606ad7c02f37e43d10a10126791dc94b0434",
        "consensus_pubkey": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "A6KRGirXFYsv5oVQz8d8YIl0Nj23bXo2jLHui72y12Bi"
        },
        "jailed": false,
        "status": 3,
        "tokens": "10061004001000000",
        "delegator_shares": "10061004001000000.000000000000000000",
        "description": {
          "moniker": "0x9DFC26A7662106EEEC5E87B20CBB690CFCE73A05"
        },
        "commission": {
          "commission_rates": {
            "rate": "0.070000000000000000",
            "max_rate": "0.100000000000000000",
            "max_change_rate": "0.010000000000000000"
          },
          "update_time": "2025-01-19T15:00:00Z"
        },
        "support_token_type": 1,
        "uptime": "99.82%",
        "apr": "36.86%"
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

#### Path Params

| Name                   | Type   | Example                                    | Required |
|------------------------|--------|--------------------------------------------|----------|
| validator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Response

- operator_address: The evm address of the validator.
- consensus_pubkey: The base64 encoded compressed public key of the validator.
- jailed: Whether the validator is jailed.
- status: The status of the validator.
  - 1: `BOND_STATUS_BONDED`
  - 2: `BOND_STATUS_UNBONDING`
  - 3: `BOND_STATUS_UNBONDED`
- tokens: The total staked tokens on the validator in `gwei`.
- delegator_shares: The total delegator shares on the validator.
- description: The description of the validator.
  - moniker: The moniker of the validator.
- commission: The commission of the validator.
- support_token_type: The support token type of the validator.
  - 0: `LOCKED`
  - 1: `UNLOCKED`
- uptime: The uptime of the validator, empty if the validator has never been bonded.
- apr: The apr of the validator, affected by the network apr and the validator's commission rate.

```json
{
  "code": 200,
  "msg": {
    "operator_address": "0xcd5faabca5bea3c5fc5e2371c7b397604720c2c2",
    "consensus_pubkey": {
      "type": "tendermint/PubKeySecp256k1",
      "value": "A/9SMxZTnh3Rq96Eygg9MfB6g82euMXhjT5nMWrhLlyf"
    },
    "jailed": false,
    "status": 3,
    "tokens": "10035524001000000",
    "delegator_shares": "10035524001000000.000000000000000000",
    "description": {
      "moniker": "0x99C28AE30CBEFEFF75E91C66692FE0BD9279B861"
    },
    "commission": {
      "commission_rates": {
        "rate": "0.070000000000000000",
        "max_rate": "0.100000000000000000",
        "max_change_rate": "0.010000000000000000"
      },
      "update_time": "2025-01-19T15:00:00Z"
    },
    "support_token_type": 0,
    "uptime": "98.64%",
    "apr": "18.43%"
  },
  "error": ""
}
```

### 4. Delegations of a Validator

[GET] `/api/staking/validators/{validator_address}/delegations`

#### Path Params

| Name                   | Type   | Example                                    | Required |
|------------------------|--------|--------------------------------------------|----------|
| validator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Query Params

| Name                   | Type   | Example | Required |
|------------------------|--------|---------|----------|
| pagination.key         | string |         | No       |
| pagination.offset      | string | 100     | No       |
| pagination.limit       | string | 10      | No       |
| pagination.count_total | string | true    | No       |
| pagination.reverse     | string | false   | No       |

#### Response

- delegation_responses: The list of delegations.
  - delegation: The delegation info.
    - delegator_address: The evm address of the delegator.
    - validator_address: The evm address of the validator.
  - balance: The balance of the delegation.
    - denom: The denom of the balance.
    - amount: The amount of the balance in `gwei`.
- pagination: The pagination info.
  - next_key: The key to query the next page.
  - total: The total number of delegations.

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

#### Path Params

| Name                   | Type   | Example                                    | Required |
|------------------------|--------|--------------------------------------------|----------|
| validator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |
| delegator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Response

- delegation_response: The delegation info.
  - delegation: The delegation info.
    - delegator_address: The evm address of the delegator.
    - validator_address: The evm address of the validator.
  - balance: The balance of the delegation.
    - denom: The denom of the balance.
    - amount: The amount of the balance in `gwei`.

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

#### Path Params

| Name                   | Type   | Example                                    | Required |
|------------------------|--------|--------------------------------------------|----------|
| validator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |
| delegator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Query Params

| Name                   | Type   | Example | Required |
|------------------------|--------|---------|----------|
| pagination.key         | string |         | No       |
| pagination.offset      | string | 100     | No       |
| pagination.limit       | string | 10      | No       |
| pagination.count_total | string | true    | No       |
| pagination.reverse     | string | false   | No       |

#### Response

- period_delegation_responses: The list of period delegations.
  - period_delegation: The period delegation info.
    - delegator_address: The evm address of the delegator.
    - validator_address: The evm address of the validator.
    - period_delegation_id: The id of the period delegation.
    - end_time: The time after which unstaking is allowed for the period delegation.
  - balance: The balance of the period delegation.
    - denom: The denom of the balance.
    - amount: The amount of the balance in `gwei`.
- pagination: The pagination info.
  - next_key: The key to query the next page.
  - total: The total number of period delegations.

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

#### Path Params

| Name                   | Type   | Example                                    | Required |
|------------------------|--------|--------------------------------------------|----------|
| validator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |
| delegator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |
| period_delegation_id   | string | 0                                          | Yes      |

#### Response

- period_delegation: The period delegation info.
  - delegator_address: The evm address of the delegator.
  - validator_address: The evm address of the validator.
  - period_delegation_id: The id of the period delegation.
  - end_time: The time after which unstaking is allowed for the period delegation.
- balance: The balance of the period delegation.
  - denom: The denom of the balance.
  - amount: The amount of the balance in `gwei`.

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

#### Path Params

| Name                   | Type   | Example                                    | Required |
|------------------------|--------|--------------------------------------------|----------|
| delegator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Query Params

| Name                   | Type   | Example | Required |
|------------------------|--------|---------|----------|
| pagination.key         | string |         | No       |
| pagination.offset      | string | 100     | No       |
| pagination.limit       | string | 10      | No       |
| pagination.count_total | string | true    | No       |
| pagination.reverse     | string | false   | No       |

#### Response

- delegation_responses: The list of delegations.
  - delegation: The delegation info.
    - delegator_address: The evm address of the delegator.
    - validator_address: The evm address of the validator.
  - balance: The balance of the delegation.
    - denom: The denom of the balance.
    - amount: The amount of the balance in `gwei`.
- pagination: The pagination info.
  - next_key: The key to query the next page.
  - total: The total number of delegations.

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

#### Path Params

| Name                   | Type   | Example                                    | Required |
|------------------------|--------|--------------------------------------------|----------|
| delegator_address      | string | 0x64a2fdc6f7cd8aa42e0bb59bf80bc47bffbe4a73 | Yes      |

#### Query Params

| Name                   | Type   | Example | Required |
|------------------------|--------|---------|----------|
| pagination.key         | string |         | No       |
| pagination.offset      | string | 100     | No       |
| pagination.limit       | string | 10      | No       |
| pagination.count_total | string | true    | No       |
| pagination.reverse     | string | false   | No       |

#### Response

- unbonding_responses: The list of unbonding delegations.
  - delegator_address: The evm address of the delegator.
  - validator_address: The evm address of the validator.
  - entries: The list of entries in the unbonding delegation.
    - creation_height: The height at which the unbonding delegation was created.
    - completion_time: The time at which the unbonding delegation will be completed.
    - initial_balance: The initial balance of the unbonding delegation.
    - balance: The current balance of the unbonding delegation, may be less than the initial balance due to slashing.
    - unbonding_id: The unique id of the unbonding delegation.
- pagination: The pagination info.
  - next_key: The key to query the next page.
  - total: The total number of unbonding delegations.

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
