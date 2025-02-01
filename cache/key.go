package cache

import "fmt"

const (
	RewardsKeyPrefix    = "rewards"
	ValidatorsKeyPrefix = "validators"

	// {prefix}_{evm_address}
	RewardsKeyFormat = "%s_%s"
	// {prefix}_{status}_{page.key}_{page.offset}_{page.limit}_{page.count_total}_{page.reverse}
	ValidatorsKeyFormat = "%s_%s_%s_%s_%s_%s_%s"
)

func RewardsKey(evmAddr string) string {
	return fmt.Sprintf(RewardsKeyFormat, RewardsKeyPrefix, evmAddr)
}

func ValidatorsKey(params map[string]string) string {
	return fmt.Sprintf(
		ValidatorsKeyFormat, ValidatorsKeyPrefix, params["status"], params["pagination.key"],
		params["pagination.offset"], params["pagination.limit"], params["pagination.count_total"], params["pagination.reverse"],
	)
}
