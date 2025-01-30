package server

import (
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/piplabs/story-indexer/db"
)

type NetworkStatus string

const (
	StatusNormal   NetworkStatus = "Normal"
	StatusDegraded NetworkStatus = "Degraded"
	StatusDown     NetworkStatus = "Down"
)

type Response struct {
	Code  int    `json:"code"`
	Msg   any    `json:"msg"`
	Error string `json:"error"`
}

type NetworkStatusData struct {
	Status        NetworkStatus `json:"status"`
	CLBlockNumber int64         `json:"consensus_block_height"`
	ELBlockNumber int64         `json:"execution_block_height"`
}

type EstimatedAPRData struct {
	APR string `json:"apr"`
}

type OperationsData struct {
	Operations []*db.Operation `json:"operations"`
	Count      int             `json:"count"`
	Total      int64           `json:"total"`
}

type RewardsData struct {
	Address          string `json:"address"`
	Amount           string `json:"amount"`
	LastUpdateHeight int64  `json:"last_update_height"`
}

type StakingValidatorData struct {
	stakingtypes.Validator
	Uptime string `json:"uptime"`
}

type StakingValidatorsData struct {
	Validators []StakingValidatorData   `json:"validators"`
	Pagination *querytypes.PageResponse `json:"pagination"`
}
