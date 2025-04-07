package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/cache"
	"github.com/piplabs/story-staking-api/db"
	"github.com/piplabs/story-staking-api/pkg/util"
)

type Interval string

const (
	IntervalOneDay     Interval = "1d"
	IntervalSevenDays  Interval = "7d"
	IntervalThirtyDays Interval = "30d"
	IntervalAllTime    Interval = "all"
)

func (s *Server) GetSystemAPRPercentage() (decimal.Decimal, error) {
	distParamsResp, err := GetDistributionParams(s.conf.Blockchain.StoryAPIEndpoint)
	if err != nil {
		return decimal.Zero, err
	}

	mintParamsResp, err := GetMintParams(s.conf.Blockchain.StoryAPIEndpoint)
	if err != nil {
		return decimal.Zero, err
	}

	inflationsPerYear, err := decimal.NewFromString(mintParamsResp.Msg.Params.InflationsPerYear)
	if err != nil {
		return decimal.Zero, err
	}

	activeValidatorsResp, err := GetStakingValidators(s.conf.Blockchain.StoryAPIEndpoint, map[string]string{
		"pagination.limit":       "100",
		"pagination.count_total": "true",
	})
	if err != nil {
		return decimal.Zero, err
	}

	bondedRewardsTokens := decimal.Zero
	for _, val := range activeValidatorsResp.Msg.Validators {
		rewardsTokens, err := decimal.NewFromString(val.RewardsTokens)
		if err != nil {
			return decimal.Zero, err
		}
		bondedRewardsTokens = bondedRewardsTokens.Add(rewardsTokens)
	}

	ubi := decimal.NewFromInt(0)
	if distParamsResp.Msg.Params.Ubi != "" {
		ubi, err = decimal.NewFromString(distParamsResp.Msg.Params.Ubi)
		if err != nil {
			return decimal.Zero, err
		}
	}

	// APR = 100% * inflations_per_year * (1 - ubi) / wighted_bonded_tokens
	aprPercentage := decimal.NewFromInt(100).
		Mul(inflationsPerYear).
		Mul(decimal.NewFromInt(1).Sub(ubi)).
		Div(bondedRewardsTokens)

	return aprPercentage, nil
}

func (s *Server) StakingParamsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingParamsHandler").Logger()

		stakingParamsResp, err := GetStakingParams(s.conf.Blockchain.StoryAPIEndpoint)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking params")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingParamsResp.Msg,
		})
	}
}
func (s *Server) NetworkStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "NetworkStatusHandler").Logger()

		clBlk, err := db.GetLatestCLBlock(s.dbOperator)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get latest cl block")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}
		isCLPaused := time.Since(clBlk.Time) > time.Minute*10

		elBlks, err := db.GetLatestELBlock(s.dbOperator, 3)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get latest el block")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}
		isELPaused := time.Since(elBlks[0].Time) > time.Minute*10

		isELCongested := true
		for i := range elBlks {
			gasUsedPercentage := decimal.NewFromInt(100).
				Mul(decimal.NewFromUint64(elBlks[i].GasUsed)).
				Div(decimal.NewFromUint64(elBlks[i].GasLimit))

			isELCongested = isELCongested && gasUsedPercentage.GreaterThanOrEqual(decimal.NewFromInt(99))
		}

		var ns NetworkStatus
		if !isCLPaused && !isELPaused && !isELCongested {
			ns = StatusNormal
		} else if isELCongested {
			ns = StatusDegraded
		} else {
			ns = StatusDown
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg: NetworkStatusData{
				Status:        ns,
				CLBlockNumber: clBlk.Height,
				ELBlockNumber: elBlks[0].Height,
			},
		})
	}
}

func (s *Server) EstimatedAPRHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "EstimatedAPRHandler").Logger()

		sysAPR, err := s.GetSystemAPRPercentage()
		if err != nil {
			logger.Error().Err(err).Msg("failed to get system apr")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  sysAPR.Truncate(2).String() + "%",
		})
	}
}

func (s *Server) OperationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "OperationsHandler").Logger()

		evmAddr := strings.ToLower(c.Param("evm_address"))
		if evmAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		pageStr := c.Query("page")
		if pageStr == "" {
			pageStr = "1"
		}
		perPageStr := c.Query("per_page")
		if perPageStr == "" {
			perPageStr = "100"
		}

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			logger.Error().Err(err).Msg("failed to parse page")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		perPage, err := strconv.Atoi(perPageStr)
		if err != nil {
			logger.Error().Err(err).Msg("failed to parse per_page")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		operations, total, err := db.GetOperations(s.dbOperator, evmAddr, page, perPage)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get operations")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg: OperationsData{
				Operations: operations,
				Count:      len(operations),
				Total:      total,
			},
		})
	}
}

func (s *Server) RewardsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "RewardsHandler").Logger()

		evmAddr := strings.ToLower(c.Param("evm_address"))
		if evmAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		// Get from cache
		cachedMsg, ok := GetCachedData[RewardsData](s.ctx, s.cacheOperator, cache.RewardsKey(evmAddr))
		if ok {
			c.JSON(http.StatusOK, Response{
				Code: http.StatusOK,
				Msg:  cachedMsg,
			})
			return
		}

		// Get from database
		rewards, err := db.GetELRewards(s.dbOperator, evmAddr)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, Response{
				Code: http.StatusOK,
				Msg: RewardsData{
					Address: evmAddr,
					Amount:  "0",
				},
			})
			return
		} else if err != nil {
			logger.Error().Err(err).Msg("failed to get rewards")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}

		msg := RewardsData{
			Address:          rewards.Address,
			Amount:           rewards.Amount,
			LastUpdateHeight: rewards.LastUpdateHeight,
		}

		// Set to cache
		_ = SetCachedData(s.ctx, s.cacheOperator, cache.RewardsKey(evmAddr), msg)

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  msg,
		})
	}
}

func (s *Server) TotalStakeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "TotalStakeHandler").Logger()

		row, err := db.GetLatestCLTotalStake(s.dbOperator)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get latest cl total stake")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg: map[string]any{
				"total_stake_amount": row.TotalStakeAmount,
				"last_updated_time":  row.CreatedAt,
			},
		})
	}
}

func (s *Server) TotalStakeHistoryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "TotalStakeHistoryHandler").Logger()

		interval := c.Query("interval")
		if interval == "" {
			interval = string(IntervalOneDay)
		}

		var (
			startTime   time.Time
			currentTime = time.Now()
		)
		switch Interval(interval) {
		case IntervalOneDay:
			startTime = currentTime.AddDate(0, 0, -1)
		case IntervalSevenDays:
			startTime = currentTime.AddDate(0, 0, -7)
		case IntervalThirtyDays:
			startTime = currentTime.AddDate(0, 0, -30)
		case IntervalAllTime:
			// no filter needed
		default:
			logger.Error().Str("interval", interval).Msg("invalid interval")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		var stakeHistory []StakeAmountData
		// Get the last amount before the period
		if Interval(interval) == IntervalAllTime {
			row, err := db.GetLatestCLTotalStakeBefore(s.dbOperator, startTime.Unix())
			if err != nil {
				logger.Error().Err(err).Msg("failed to get latest cl total stake before period")
				c.JSON(http.StatusOK, Response{
					Code:  http.StatusInternalServerError,
					Error: ErrInternalDataServiceError.Error(),
				})
				return
			}
			stakeHistory = append(stakeHistory, StakeAmountData{
				TotalStakeAmount: row.TotalStakeAmount,
				Timestamp:        row.CreatedAt,
			})
		}
		// Get all amount updates after the start time
		if Interval(interval) == IntervalAllTime {
			rows, err := db.GetCLTotalStakes(s.dbOperator)
			if err != nil {
				logger.Error().Err(err).Msg("failed to get all cl total stakes")
				c.JSON(http.StatusOK, Response{
					Code:  http.StatusInternalServerError,
					Error: ErrInternalDataServiceError.Error(),
				})
				return
			}
			for _, row := range rows {
				stakeHistory = append(stakeHistory, StakeAmountData{
					TotalStakeAmount: row.TotalStakeAmount,
					Timestamp:        row.CreatedAt,
				})
			}
		} else {
			rows, err := db.GetCLTotalStakesAfter(s.dbOperator, startTime.Unix())
			if err != nil {
				logger.Error().Err(err).Msg("failed to get cl total stakes within period")
				c.JSON(http.StatusOK, Response{
					Code:  http.StatusInternalServerError,
					Error: ErrInternalDataServiceError.Error(),
				})
				return
			}
			for _, row := range rows {
				stakeHistory = append(stakeHistory, StakeAmountData{
					TotalStakeAmount: row.TotalStakeAmount,
					Timestamp:        row.CreatedAt,
				})
			}
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg: map[string]any{
				"total_stake_amount_history": stakeHistory,
			},
		})
	}
}

func (s *Server) StakingPoolHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingPoolHandler").Logger()

		stakingPoolResp, err := GetStakingPool(s.conf.Blockchain.StoryAPIEndpoint)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking pool")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingPoolResp.Msg,
		})
	}
}

func (s *Server) StakingValidatorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingValidatorsHandler").Logger()

		params := ParsePaginationParams(c)
		if status := c.Query("status"); status != "" {
			params["status"] = status
		}

		// Get from cache
		cachedMsg, ok := GetCachedData[StakingValidatorsData](s.ctx, s.cacheOperator, cache.ValidatorsKey(params))
		if ok {
			c.JSON(http.StatusOK, Response{
				Code: http.StatusOK,
				Msg:  cachedMsg,
			})
			return
		}

		sysAPR, err := s.GetSystemAPRPercentage()
		if err != nil {
			logger.Error().Err(err).Msg("failed to get system apr")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		// Query from API and database
		stakingValidatorsResp, err := GetStakingValidators(s.conf.Blockchain.StoryAPIEndpoint, params)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking validators")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		valAddrs := make([]string, 0, len(stakingValidatorsResp.Msg.Validators))
		for _, val := range stakingValidatorsResp.Msg.Validators {
			valAddrs = append(valAddrs, strings.ToLower(val.OperatorAddress))
		}

		valVotes, err := db.GetCLValidatorsVotes(s.dbOperator, valAddrs...)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get cl uptimes")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}

		clUptimesMap := make(map[string]string)
		for valAddr, votes := range valVotes {
			clUptimesMap[strings.ToLower(valAddr)] = decimal.NewFromInt(100).
				Mul(decimal.NewFromInt(votes)).
				Div(decimal.NewFromInt(util.UptimeWindow)).
				Truncate(2).String() + "%"
		}

		validators := make([]StakingValidatorData, 0, len(stakingValidatorsResp.Msg.Validators))
		for _, val := range stakingValidatorsResp.Msg.Validators {
			commissionRate, err := decimal.NewFromString(val.Commission.CommissionRates.Rate)
			if err != nil {
				logger.Error().Err(err).Str("validator", val.OperatorAddress).Msg("failed to parse commission rate")
				c.JSON(http.StatusOK, Response{
					Code:  http.StatusInternalServerError,
					Error: ErrInternalDataServiceError.Error(),
				})
				return
			}

			valAPR := sysAPR.Mul(decimal.NewFromInt(1).Sub(commissionRate))
			// Locked token type has 0.5x APR
			if val.SupportTokenType == 0 {
				valAPR = valAPR.Div(decimal.NewFromInt(2))
			}

			validators = append(validators, StakingValidatorData{
				ValidatorInfo: val,
				Uptime:        clUptimesMap[strings.ToLower(val.OperatorAddress)],
				APR:           valAPR.Truncate(2).String() + "%",
			})
		}

		msg := StakingValidatorsData{
			Validators: validators,
			Pagination: stakingValidatorsResp.Msg.Pagination,
		}

		// Set to cache
		_ = SetCachedData(s.ctx, s.cacheOperator, cache.ValidatorsKey(params), msg)

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  msg,
		})
	}
}

func (s *Server) StakingValidatorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingValidatorHandler").Logger()

		valAddr := strings.ToLower(c.Param("validator_address"))
		if valAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		sysAPR, err := s.GetSystemAPRPercentage()
		if err != nil {
			logger.Error().Err(err).Msg("failed to get system apr")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		// Query from API and database
		stakingValidatorResp, err := GetStakingValidator(s.conf.Blockchain.StoryAPIEndpoint, valAddr)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking validator")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		val := stakingValidatorResp.Msg.Validator

		valVotes, err := db.GetCLValidatorsVotes(s.dbOperator, valAddr)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get cl uptimes")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}

		clUptimesMap := make(map[string]string)
		for valAddr, votes := range valVotes {
			clUptimesMap[strings.ToLower(valAddr)] = decimal.NewFromInt(100).
				Mul(decimal.NewFromInt(votes)).
				Div(decimal.NewFromInt(util.UptimeWindow)).
				Truncate(2).String() + "%"
		}

		commissionRate, err := decimal.NewFromString(val.Commission.CommissionRates.Rate)
		if err != nil {
			logger.Error().Err(err).Str("validator", val.OperatorAddress).Msg("failed to parse commission rate")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalDataServiceError.Error(),
			})
			return
		}

		valAPR := sysAPR.Mul(decimal.NewFromInt(1).Sub(commissionRate))
		// Locked token type has 0.5x APR
		if val.SupportTokenType == 0 {
			valAPR = valAPR.Div(decimal.NewFromInt(2))
		}

		msg := StakingValidatorData{
			ValidatorInfo: val,
			Uptime:        clUptimesMap[strings.ToLower(val.OperatorAddress)],
			APR:           valAPR.Truncate(2).String() + "%",
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  msg,
		})
	}
}

func (s *Server) StakingValidatorDelegationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingValidatorsHandler").Logger()

		valAddr := strings.ToLower(c.Param("validator_address"))
		if valAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		params := ParsePaginationParams(c)

		stakingValidatorDelegationsResp, err := GetStakingValidatorDelegations(s.conf.Blockchain.StoryAPIEndpoint, valAddr, params)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking validator delegations")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingValidatorDelegationsResp.Msg,
		})
	}
}

func (s *Server) StakingDelegationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingDelegationHandler").Logger()

		valAddr := strings.ToLower(c.Param("validator_address"))
		delAddr := strings.ToLower(c.Param("delegator_address"))
		if valAddr == "" || delAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		stakingDelegationResp, err := GetStakingDelegation(s.conf.Blockchain.StoryAPIEndpoint, valAddr, delAddr)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking delegation")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingDelegationResp.Msg,
		})
	}
}

func (s *Server) StakingValidatorDelegatorPeriodDelegationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingValidatorDelegatorPeriodDelegationsHandler").Logger()

		valAddr := strings.ToLower(c.Param("validator_address"))
		delAddr := strings.ToLower(c.Param("delegator_address"))
		if valAddr == "" || delAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		params := ParsePaginationParams(c)

		stakingValidatorDelegatorPeriodDelegationsResp, err := GetStakingValidatorDelegatorPeriodDelegations(s.conf.Blockchain.StoryAPIEndpoint, valAddr, delAddr, params)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking validator delegator period delegations")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingValidatorDelegatorPeriodDelegationsResp.Msg,
		})
	}
}

func (s *Server) StakingValidatorDelegatorPeriodDelegationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingValidatorDelegatorPeriodDelegationHandler").Logger()

		valAddr := strings.ToLower(c.Param("validator_address"))
		delAddr := strings.ToLower(c.Param("delegator_address"))
		delID := c.Param("period_delegation_id")
		if valAddr == "" || delAddr == "" || delID == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		stakingValidatorDelegatorPeriodDelegationResp, err := GetStakingValidatorDelegatorPeriodDelegation(s.conf.Blockchain.StoryAPIEndpoint, valAddr, delAddr, delID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking validator delegator period delegation")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingValidatorDelegatorPeriodDelegationResp.Msg,
		})
	}
}

func (s *Server) StakingDelegatorDelegationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingDelegatorDelegationsHandler").Logger()

		delAddr := strings.ToLower(c.Param("delegator_address"))
		if delAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		params := ParsePaginationParams(c)

		stakingDelegatorDelegationsResp, err := GetStakingDelegatorDelegations(s.conf.Blockchain.StoryAPIEndpoint, delAddr, params)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking delegator delegations")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingDelegatorDelegationsResp.Msg,
		})
	}
}

func (s *Server) StakingDelegatorUnbondingDelegationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := log.With().Str("handler", "StakingDelegatorUnbondingDelegationsHandler").Logger()

		delAddr := strings.ToLower(c.Param("delegator_address"))
		if delAddr == "" {
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusBadRequest,
				Error: ErrInvalidParameter.Error(),
			})
			return
		}

		params := ParsePaginationParams(c)

		stakingDelegatorUnbondingDelegationsResp, err := GetStakingDelegatorUnbondingDelegations(s.conf.Blockchain.StoryAPIEndpoint, delAddr, params)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get staking delegator unbonding delegations")
			c.JSON(http.StatusOK, Response{
				Code:  http.StatusInternalServerError,
				Error: ErrInternalAPIServiceError.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  stakingDelegatorUnbondingDelegationsResp.Msg,
		})
	}
}
