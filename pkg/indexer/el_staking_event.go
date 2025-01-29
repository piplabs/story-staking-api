package indexer

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/piplabs/story-indexer/db"
	"github.com/piplabs/story-indexer/pkg/indexer/contract/iptokenstaking"
	"github.com/piplabs/story-indexer/pkg/util"
)

const (
	TypeSetOperator               = "SetOperator"
	TypeUnsetOperator             = "UnsetOperator"
	TypeSetWithdrawalAddress      = "SetWithdrawalAddress"
	TypeSetRewardAddress          = "SetRewardAddress"
	TypeUpdateValidatorCommission = "UpdateValidatorCommission"
	TypeCreateValidator           = "CreateValidator"
	TypeStake                     = "Stake"
	TypeStakeOnBehalf             = "StakeOnBehalf"
	TypeRedelegate                = "Redelegate"
	TypeRedelegateOnBehalf        = "RedelegateOnBehalf"
	TypeUnstake                   = "Unstake"
	TypeUnstakeOnBehalf           = "UnstakeOnBehalf"
	TypeUnjail                    = "Unjail"
	TypeUnjailOnBehalf            = "UnjailOnBehalf"
)

type ELStakingEventIndexer struct {
	ctx context.Context

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	ethClient     *ethclient.Client
	elEventFilter *iptokenstaking.IPTokenStakingFilterer
}

func NewELStakingEventIndexer(ctx context.Context, dbOperator *gorm.DB, cacheOperator *redis.Client, rpcEndpoint string) (*ELStakingEventIndexer, error) {
	ethClient, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return nil, err
	}

	elEventFilter, err := iptokenstaking.NewIPTokenStakingFilterer(iptokenstaking.ContractAddress, ethClient)
	if err != nil {
		return nil, err
	}

	return &ELStakingEventIndexer{
		ctx: ctx,

		dbOperator:    dbOperator,
		cacheOperator: cacheOperator,

		ethClient:     ethClient,
		elEventFilter: elEventFilter,
	}, nil
}

func (e *ELStakingEventIndexer) Name() string {
	return "el_staking_event"
}

func (e *ELStakingEventIndexer) Run() {
	log.Info().Str("indexer", e.Name()).Msg("Start indexing")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			indexPoint, err := db.GetIndexPoint(e.dbOperator, e.Name())
			if err != nil {
				log.Error().Err(err).Str("indexer", e.Name()).Msg("get index point failed")
				continue
			}

			latestBlkNum, err := e.ethClient.BlockNumber(e.ctx)
			if err != nil {
				log.Error().Err(err).Str("indexer", e.Name()).Msg("get latest el block failed")
				continue
			}

			if indexPoint.BlockHeight+10 > int64(latestBlkNum) {
				continue
			}

			if err := e.index(indexPoint.BlockHeight+1, int64(latestBlkNum)); err != nil {
				log.Error().Err(err).
					Str("indexer", e.Name()).
					Int64("from", indexPoint.BlockHeight).
					Uint64("to", latestBlkNum).
					Msg("index el staking event failed")
			}
		}
	}
}

func (e *ELStakingEventIndexer) index(from, to int64) error {
	start := from

	for start <= to {
		end := min(start+100, to)

		stakingEvents, err := e.getStakingEvents(start, end)
		if err != nil {
			return err
		}

		if err := db.BatchCreateELStakingEvents(e.dbOperator, e.Name(), stakingEvents, end); err != nil {
			return err
		}

		start = end + 1
	}

	return nil
}

func (e *ELStakingEventIndexer) getStakingEvents(from, to int64) ([]*db.ELStakingEvent, error) {
	fromBlock, toBlock := uint64(from), uint64(to)

	var elStakingEvents []*db.ELStakingEvent
	// SetOperator event.
	setOperatorEvents, err := e.elEventFilter.FilterSetOperator(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer setOperatorEvents.Close()

	for setOperatorEvents.Next() {
		ev := setOperatorEvents.Event

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:      hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight: int64(ev.Raw.BlockNumber),
			EventType:   TypeSetOperator,
			Address:     strings.ToLower(ev.Delegator.Hex()),
			DstAddress:  strings.ToLower(ev.Operator.Hex()),
		})
	}

	// UnsetOperator event.
	unsetOperatorEvents, err := e.elEventFilter.FilterUnsetOperator(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer unsetOperatorEvents.Close()

	for unsetOperatorEvents.Next() {
		ev := unsetOperatorEvents.Event

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:      hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight: int64(ev.Raw.BlockNumber),
			EventType:   TypeUnsetOperator,
			Address:     strings.ToLower(ev.Delegator.Hex()),
		})
	}

	// SetWithdrawalAddress event.
	setWithdrawalAddressEvents, err := e.elEventFilter.FilterSetWithdrawalAddress(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer setWithdrawalAddressEvents.Close()

	for setWithdrawalAddressEvents.Next() {
		ev := setWithdrawalAddressEvents.Event

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:      hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight: int64(ev.Raw.BlockNumber),
			EventType:   TypeSetWithdrawalAddress,
			Address:     strings.ToLower(ev.Delegator.Hex()),
			DstAddress:  strings.ToLower(common.BytesToAddress(ev.ExecutionAddress[:]).Hex()),
		})
	}

	// SetRewardAddress event.
	setRewardAddressEvents, err := e.elEventFilter.FilterSetRewardAddress(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer setRewardAddressEvents.Close()

	for setRewardAddressEvents.Next() {
		ev := setRewardAddressEvents.Event

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:      hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight: int64(ev.Raw.BlockNumber),
			EventType:   TypeSetRewardAddress,
			Address:     strings.ToLower(ev.Delegator.Hex()),
			DstAddress:  strings.ToLower(common.BytesToAddress(ev.ExecutionAddress[:]).Hex()),
		})
	}

	// UpdateValidatorCommission event.
	updateValidatorCommissionEvents, err := e.elEventFilter.FilterUpdateValidatorCommission(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer updateValidatorCommissionEvents.Close()

	for updateValidatorCommissionEvents.Next() {
		ev := updateValidatorCommissionEvents.Event

		evmAddr, err := util.CmpPubKeyToEVMAddress(ev.ValidatorCmpPubkey)
		if err != nil {
			return nil, err
		}

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:      hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight: int64(ev.Raw.BlockNumber),
			EventType:   TypeUpdateValidatorCommission,
			Address:     strings.ToLower(evmAddr.Hex()),
		})
	}

	// CreateValidator event.
	createValidatorEvents, err := e.elEventFilter.FilterCreateValidator(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer createValidatorEvents.Close()

	for createValidatorEvents.Next() {
		ev := createValidatorEvents.Event

		valAddr := strings.ToLower(ev.OperatorAddress.Hex())

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:              hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight:         int64(ev.Raw.BlockNumber),
			EventType:           TypeCreateValidator,
			Address:             valAddr,
			DstValidatorAddress: valAddr,
		})
	}

	// Deposit event.
	depositEvents, err := e.elEventFilter.FilterDeposit(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer depositEvents.Close()

	for depositEvents.Next() {
		ev := depositEvents.Event

		eventType := TypeStake
		if ev.OperatorAddress.Hex() != ev.Delegator.Hex() {
			eventType = TypeStakeOnBehalf
		}

		valAddr, err := util.CmpPubKeyToEVMAddress(ev.ValidatorCmpPubkey)
		if err != nil {
			return nil, err
		}

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:              hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight:         int64(ev.Raw.BlockNumber),
			EventType:           eventType,
			Address:             strings.ToLower(ev.OperatorAddress.Hex()),
			DstValidatorAddress: strings.ToLower(valAddr.Hex()),
		})
	}

	// Redelegate event.
	redelegateEvents, err := e.elEventFilter.FilterRedelegate(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer redelegateEvents.Close()

	for redelegateEvents.Next() {
		ev := redelegateEvents.Event

		eventType := TypeRedelegate
		if ev.OperatorAddress.Hex() != ev.Delegator.Hex() {
			eventType = TypeRedelegateOnBehalf
		}

		srcValAddr, err := util.CmpPubKeyToEVMAddress(ev.ValidatorSrcCmpPubkey)
		if err != nil {
			return nil, err
		}

		dstValAddr, err := util.CmpPubKeyToEVMAddress(ev.ValidatorDstCmpPubkey)
		if err != nil {
			return nil, err
		}

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:              hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight:         int64(ev.Raw.BlockNumber),
			EventType:           eventType,
			Address:             strings.ToLower(ev.OperatorAddress.Hex()),
			SrcValidatorAddress: strings.ToLower(srcValAddr.Hex()),
			DstValidatorAddress: strings.ToLower(dstValAddr.Hex()),
		})
	}

	// Withdraw event.
	withdrawEvents, err := e.elEventFilter.FilterWithdraw(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer withdrawEvents.Close()

	for withdrawEvents.Next() {
		ev := withdrawEvents.Event

		eventType := TypeUnstake
		if ev.OperatorAddress.Hex() != ev.Delegator.Hex() {
			eventType = TypeUnstakeOnBehalf
		}

		valAddr, err := util.CmpPubKeyToEVMAddress(ev.ValidatorCmpPubkey)
		if err != nil {
			return nil, err
		}

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:              hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight:         int64(ev.Raw.BlockNumber),
			EventType:           eventType,
			Address:             strings.ToLower(ev.OperatorAddress.Hex()),
			DstValidatorAddress: strings.ToLower(valAddr.Hex()),
		})
	}

	// Unjail event.
	unjailEvents, err := e.elEventFilter.FilterUnjail(&bind.FilterOpts{
		Context: e.ctx,
		Start:   fromBlock,
		End:     &toBlock,
	})
	if err != nil {
		return nil, err
	}
	defer unjailEvents.Close()

	for unjailEvents.Next() {
		ev := unjailEvents.Event

		valAddr, err := util.CmpPubKeyToEVMAddress(ev.ValidatorCmpPubkey)
		if err != nil {
			return nil, err
		}

		eventType := TypeUnjail
		if ev.Unjailer.Hex() != valAddr.Hex() {
			eventType = TypeUnjailOnBehalf
		}

		elStakingEvents = append(elStakingEvents, &db.ELStakingEvent{
			TxHash:              hex.EncodeToString(ev.Raw.TxHash.Bytes()),
			BlockHeight:         int64(ev.Raw.BlockNumber),
			EventType:           eventType,
			Address:             strings.ToLower(ev.Unjailer.Hex()),
			DstValidatorAddress: strings.ToLower(valAddr.Hex()),
		})
	}

	return elStakingEvents, nil
}
