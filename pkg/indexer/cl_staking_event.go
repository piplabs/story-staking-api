package indexer

import (
	"context"
	"strconv"
	"time"

	lightprovider "github.com/cometbft/cometbft/light/provider"
	lighthttp "github.com/cometbft/cometbft/light/provider/http"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-indexer/db"
)

var _ Indexer = (*CLStakinEventIndexer)(nil)

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

type CLStakinEventIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	cometClient      *comethttp.HTTP
	lightCometClient lightprovider.Provider
}

func NewCLStakinEventIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, chainID, rpcEndpoint string) (*CLStakinEventIndexer, error) {
	cometClient, err := comethttp.New(rpcEndpoint, "")
	if err != nil {
		return nil, err
	}

	lightCometClient, err := lighthttp.New(chainID, rpcEndpoint)
	if err != nil {
		return nil, err
	}

	return &CLStakinEventIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		cometClient:      cometClient,
		lightCometClient: lightCometClient,
	}, nil
}

func (c *CLStakinEventIndexer) Name() string {
	return "cl_staking_event"
}

func (c *CLStakinEventIndexer) Run() {
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

func (c *CLStakinEventIndexer) index(from, to int64) error {
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

	if len(clStakingEvents) > 0 {
		if err := db.BatchCreateCLStakingEvents(c.dbOperator, c.Name(), clStakingEvents, to); err != nil {
			return err
		}
	}

	return nil
}

func (c *CLStakinEventIndexer) getBlockEvents(blkno int64) ([]*db.CLStakingEvent, error) {
	blockResults, err := c.cometClient.BlockResults(c.ctx, &blkno)
	if err != nil {
		return nil, err
	}

	stakingCLEvents := make([]*db.CLStakingEvent, 0)
	for _, e := range blockResults.FinalizeBlockEvents {
		if _, ok := EventSet[e.Type]; !ok {
			continue
		}

		attrMap := attrArray2Map(e.Attributes)

		errCode, exists := attrMap[AttributeKeyErrorCode]

		amount := int64(0)
		if amountStr, exists := attrMap[AttributeKeyAmount]; exists {
			amount, err = strconv.ParseInt(amountStr, 10, 64)
			if err != nil {
				return nil, err
			}
		}

		stakingCLEvents = append(stakingCLEvents, &db.CLStakingEvent{
			ELTxHash:    attrMap[AttributeKeyTxHash],
			BlockHeight: blkno,
			StatusOK:    !exists,
			ErrorCode:   errCode,
			Amount:      amount,
		})
	}

	return stakingCLEvents, nil
}
