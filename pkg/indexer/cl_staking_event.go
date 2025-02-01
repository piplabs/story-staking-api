package indexer

import (
	"context"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	lightprovider "github.com/cometbft/cometbft/light/provider"
	lighthttp "github.com/cometbft/cometbft/light/provider/http"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

var _ Indexer = (*CLStakingEventIndexer)(nil)

const (
	EventTypeSetOperatorFailure               = "set_operator_failure"
	EventTypeUnsetOperatorFailure             = "unset_operator_failure"
	EventTypeSetWithdrawalAddressFailure      = "set_withdrawal_address_failure"
	EventTypeSetRewardAddressFailure          = "set_reward_address_failure"
	EventTypeUpdateValidatorCommissionFailure = "update_validator_commission_failure"
	EventTypeCreateValidatorFailure           = "create_validator_failure"
	EventTypeDelegateFailure                  = "delegate_failure"
	EventTypeRedelegateFailure                = "redelegate_failure"
	EventTypeUndelegateFailure                = "undelegate_failure"
	EventTypeUnjailFailure                    = "unjail_failure"

	EventTypeSetOperatorSuccess               = "set_operator_success"
	EventTypeUnsetOperatorSuccess             = "unset_operator_success"
	EventTypeSetWithdrawalAddressSuccess      = "set_withdrawal_address_success"
	EventTypeSetRewardAddressSuccess          = "set_reward_address_success"
	EventTypeUpdateValidatorCommissionSuccess = "update_validator_commission_success"
	EventTypeCreateValidatorSuccess           = "create_validator_success"
	EventTypeDelegateSuccess                  = "delegate_success"
	EventTypeRedelegateSuccess                = "redelegate_success"
	EventTypeUndelegateSuccess                = "undelegate_success"
	EventTypeUnjailSuccess                    = "unjail_success"

	AttributeKeyErrorCode          = "error_code"
	AttributeKeyTxHash             = "tx_hash"
	AttributeKeyValidatorCmpPubKey = "validator_cmp_pubkey"
	AttributeKeyAmount             = "amount"
)

var EventSet = map[string]struct{}{
	EventTypeUpdateValidatorCommissionSuccess: {},
	EventTypeUpdateValidatorCommissionFailure: {},
	EventTypeSetWithdrawalAddressSuccess:      {},
	EventTypeSetWithdrawalAddressFailure:      {},
	EventTypeSetRewardAddressSuccess:          {},
	EventTypeSetRewardAddressFailure:          {},
	EventTypeSetOperatorSuccess:               {},
	EventTypeSetOperatorFailure:               {},
	EventTypeUnsetOperatorSuccess:             {},
	EventTypeUnsetOperatorFailure:             {},
	EventTypeCreateValidatorSuccess:           {},
	EventTypeCreateValidatorFailure:           {},
	EventTypeDelegateSuccess:                  {},
	EventTypeDelegateFailure:                  {},
	EventTypeRedelegateSuccess:                {},
	EventTypeRedelegateFailure:                {},
	EventTypeUndelegateSuccess:                {},
	EventTypeUndelegateFailure:                {},
	EventTypeUnjailSuccess:                    {},
	EventTypeUnjailFailure:                    {},
}

type CLStakingEventIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	cometClient      *comethttp.HTTP
	lightCometClient lightprovider.Provider
}

func NewCLStakingEventIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, chainID, rpcEndpoint string) (*CLStakingEventIndexer, error) {
	cometClient, err := comethttp.New(rpcEndpoint, "")
	if err != nil {
		return nil, err
	}

	lightCometClient, err := lighthttp.New(chainID, rpcEndpoint)
	if err != nil {
		return nil, err
	}

	return &CLStakingEventIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		cometClient:      cometClient,
		lightCometClient: lightCometClient,
	}, nil
}

func (c *CLStakingEventIndexer) Name() string {
	return "cl_staking_event"
}

func (c *CLStakingEventIndexer) Run() {
	log.Info().Str("indexer", c.Name()).Msg("Start indexing")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			indexPoint, err := db.GetIndexPoint(c.dbOperator, c.Name())
			if err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("get index point failed")
				continue
			}

			latestBlk, err := c.lightCometClient.LightBlock(c.ctx, 0)
			if err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("get latest cl block failed")
				continue
			}

			if indexPoint.BlockHeight+10 > latestBlk.Height {
				continue
			}

			if err := c.index(indexPoint.BlockHeight+1, latestBlk.Height); err != nil {
				log.Error().Err(err).
					Str("indexer", c.Name()).
					Int64("from", indexPoint.BlockHeight).
					Int64("to", latestBlk.Height).
					Msg("index cl staking event failed")
			}
		}
	}
}

func (c *CLStakingEventIndexer) index(from, to int64) error {
	var clStakingEvents []*db.CLStakingEvent

	for i := from; i <= to; i++ {
		events, err := c.getBlockEvents(i)
		if err != nil {
			return err
		}

		clStakingEvents = append(clStakingEvents, events...)

		if len(clStakingEvents) > 100 {
			if err := db.BatchCreateCLStakingEvents(c.dbOperator, c.Name(), clStakingEvents, i); err != nil {
				return err
			}
			clStakingEvents = make([]*db.CLStakingEvent, 0)
		}
	}

	// Handle remaining entries, even if there are no entries, we also need to update the index point.
	if err := db.BatchCreateCLStakingEvents(c.dbOperator, c.Name(), clStakingEvents, to); err != nil {
		return err
	}

	return nil
}

func (c *CLStakingEventIndexer) getBlockEvents(blkno int64) ([]*db.CLStakingEvent, error) {
	blockResults, err := c.cometClient.BlockResults(c.ctx, &blkno)
	if err != nil {
		return nil, err
	}

	blockEvents := make([]abcitypes.Event, 0)
	for _, tr := range blockResults.TxsResults {
		blockEvents = append(blockEvents, tr.Events...)
	}
	blockEvents = append(blockEvents, blockResults.FinalizeBlockEvents...)

	stakingCLEvents := make([]*db.CLStakingEvent, 0)
	for _, e := range blockEvents {
		if _, ok := EventSet[e.Type]; !ok {
			continue
		}

		attrMap := attrArray2Map(e.Attributes)

		errCode, exists := attrMap[AttributeKeyErrorCode]

		stakingCLEvents = append(stakingCLEvents, &db.CLStakingEvent{
			ELTxHash:    "0x" + attrMap[AttributeKeyTxHash],
			BlockHeight: blkno,
			StatusOK:    !exists,
			ErrorCode:   errCode,
			Amount:      attrMap[AttributeKeyAmount],
		})
	}

	return stakingCLEvents, nil
}
