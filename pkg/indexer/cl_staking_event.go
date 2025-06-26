package indexer

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
	"github.com/piplabs/story-staking-api/pkg/util"
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
	AttributeKeySenderAddress      = "sender_address"
	AttributeKeyDelegatorAddress   = "delegator_addr"
)

var Event2Type = map[string]string{
	EventTypeUpdateValidatorCommissionSuccess: TypeUpdateValidatorCommission,
	EventTypeUpdateValidatorCommissionFailure: TypeUpdateValidatorCommission,
	EventTypeSetWithdrawalAddressSuccess:      TypeSetWithdrawalAddress,
	EventTypeSetWithdrawalAddressFailure:      TypeSetWithdrawalAddress,
	EventTypeSetRewardAddressSuccess:          TypeSetRewardAddress,
	EventTypeSetRewardAddressFailure:          TypeSetRewardAddress,
	EventTypeSetOperatorSuccess:               TypeSetOperator,
	EventTypeSetOperatorFailure:               TypeSetOperator,
	EventTypeUnsetOperatorSuccess:             TypeUnsetOperator,
	EventTypeUnsetOperatorFailure:             TypeUnsetOperator,
	EventTypeCreateValidatorSuccess:           TypeCreateValidator,
	EventTypeCreateValidatorFailure:           TypeCreateValidator,
	EventTypeDelegateSuccess:                  TypeStake,
	EventTypeDelegateFailure:                  TypeStake,
	EventTypeRedelegateSuccess:                TypeRedelegate,
	EventTypeRedelegateFailure:                TypeRedelegate,
	EventTypeUndelegateSuccess:                TypeUnstake,
	EventTypeUndelegateFailure:                TypeUnstake,
	EventTypeUnjailSuccess:                    TypeUnjail,
	EventTypeUnjailFailure:                    TypeUnjail,
}

var EventType2Behalf = map[string]string{
	TypeStake:      TypeStakeOnBehalf,
	TypeRedelegate: TypeRedelegateOnBehalf,
	TypeUnstake:    TypeUnstakeOnBehalf,
	TypeUnjail:     TypeUnjailOnBehalf,
}

type CLStakingEventIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	cometClient *comethttp.HTTP
}

func NewCLStakingEventIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, rpcEndpoint string) (*CLStakingEventIndexer, error) {
	cometClient, err := comethttp.New(rpcEndpoint, "")
	if err != nil {
		return nil, err
	}

	return &CLStakingEventIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		cometClient: cometClient,
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

			latestBlk, err := c.cometClient.Block(c.ctx, nil)
			if err != nil {
				log.Error().Err(err).Str("indexer", c.Name()).Msg("get latest cl block failed")
				continue
			}

			if indexPoint.BlockHeight+10 > latestBlk.Block.Height {
				continue
			}

			if err := c.index(indexPoint.BlockHeight+1, latestBlk.Block.Height); err != nil {
				log.Error().Err(err).
					Str("indexer", c.Name()).
					Int64("from", indexPoint.BlockHeight+1).
					Int64("to", latestBlk.Block.Height).
					Msg("index cl staking event failed")
			}
		}
	}
}

func (c *CLStakingEventIndexer) index(from, to int64) error {
	start := from

	for start <= to {
		end := min(start+100, to)

		stakingEvents, err := c.getStakingEvents(start, end)
		if err != nil {
			return err
		}

		if err := db.BatchCreateCLStakingEvents(c.dbOperator, c.Name(), stakingEvents, end); err != nil {
			return err
		}

		start = end + 1
	}

	return nil
}

func (c *CLStakingEventIndexer) getStakingEvents(from, to int64) ([]*db.CLStakingEvent, error) {
	stakingCLEvents := make([]*db.CLStakingEvent, 0)

	for blkno := from; blkno <= to; blkno++ {
		blockResults, err := c.cometClient.BlockResults(c.ctx, &blkno)
		if err != nil {
			return nil, err
		}

		blockEvents := make([]abcitypes.Event, 0)
		for _, tr := range blockResults.TxsResults {
			blockEvents = append(blockEvents, tr.Events...)
		}
		blockEvents = append(blockEvents, blockResults.FinalizeBlockEvents...)

		for _, e := range blockEvents {
			eventType, ok := Event2Type[e.Type]
			if !ok {
				continue
			}

			attrMap := attrArray2Map(e.Attributes)

			// Check if it's a unbehalf txn
			switch eventType {
			case TypeStake, TypeRedelegate, TypeUnstake:
				delAddr, ok := attrMap[AttributeKeyDelegatorAddress]
				if !ok {
					return nil, fmt.Errorf("event %s: delegator address not found", eventType)
				}
				senderAddr, ok := attrMap[AttributeKeySenderAddress]
				if !ok {
					return nil, fmt.Errorf("event %s: sender address not found", eventType)
				}

				if !strings.EqualFold(delAddr, senderAddr) {
					eventType = EventType2Behalf[eventType]
				}
			case TypeUnjail:
				valCmpPubKey, ok := attrMap[AttributeKeyValidatorCmpPubKey]
				if !ok {
					return nil, fmt.Errorf("event %s: validator compressed key not found", eventType)
				}
				senderAddr, ok := attrMap[AttributeKeySenderAddress]
				if !ok {
					return nil, fmt.Errorf("event %s: sender address not found", eventType)
				}

				valCmpPubKeyBytes, err := hex.DecodeString(valCmpPubKey)
				if err != nil {
					return nil, fmt.Errorf("decode validator compressed key from event %s failed: %w", eventType, err)
				}
				valAddr, err := util.CmpPubKeyToEVMAddress(valCmpPubKeyBytes)
				if err != nil {
					return nil, fmt.Errorf("convert validator compressed key to address from event %s failed: %w", eventType, err)
				}

				if !strings.EqualFold(valAddr.String(), senderAddr) {
					eventType = EventType2Behalf[eventType]
				}
			}

			errCode, exists := attrMap[AttributeKeyErrorCode]

			stakingCLEvents = append(stakingCLEvents, &db.CLStakingEvent{
				ELTxHash:    "0x" + attrMap[AttributeKeyTxHash],
				EventType:   eventType,
				BlockHeight: blkno,
				StatusOK:    !exists,
				ErrorCode:   errCode,
				Amount:      attrMap[AttributeKeyAmount],
			})
		}
	}

	return stakingCLEvents, nil
}
