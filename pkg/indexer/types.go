package indexer

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

const (
	WithdrawalTypeUnstake = 0
	WithdrawalTypeReward  = 1
	WithdrawalTypeUBI     = 2
)
