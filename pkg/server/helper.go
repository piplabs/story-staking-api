package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/piplabs/story-staking-api/pkg/metrics"
)

const (
	TokenTypeLocked   = 0
	TokenTypeUnlocked = 1
)

type QueryResponse[T any] struct {
	Code  int    `json:"code"`
	Msg   T      `json:"msg"`
	Error string `json:"error"`
}

type Pagination struct {
	NextKey string `json:"next_key"`
	Total   string `json:"total"`
}

type DistributionParamsResponse struct {
	Params struct {
		Ubi string `json:"ubi"`
	} `json:"params"`
}

type MintParamsResponse struct {
	Params struct {
		MintDenom         string `json:"mint_denom"`
		InflationsPerYear string `json:"inflations_per_year"`
		BlocksPerYear     string `json:"blocks_per_year"`
	} `json:"params"`
}

type StakingParamsResponse struct {
	Params struct {
		UnbondingTime     string `json:"unbonding_time"`
		MaxValidators     int    `json:"max_validators"`
		MaxEntries        int    `json:"max_entries"`
		HistoricalEntries int    `json:"historical_entries"`
		BondDenom         string `json:"bond_denom"`
		MinCommissionRate string `json:"min_commission_rate"`
		MinDelegation     string `json:"min_delegation"`
		Periods           []struct {
			PeriodType        int    `json:"period_type"`
			Duration          string `json:"duration"`
			RewardsMultiplier string `json:"rewards_multiplier"`
		} `json:"periods"`
		TokenTypes []struct {
			TokenType         int    `json:"token_type"`
			RewardsMultiplier string `json:"rewards_multiplier"`
		} `json:"token_types"`
		SingularityHeight string `json:"singularity_height"`
	} `json:"params"`
}

type StakingPoolResponse struct {
	Pool struct {
		NotBondedTokens string `json:"not_bonded_tokens"`
		BondedTokens    string `json:"bonded_tokens"`
	} `json:"pool"`
}

type ValidatorInfo struct {
	OperatorAddress string `json:"operator_address"`
	ConsensusPubKey struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"consensus_pubkey"`
	Jailed          bool   `json:"jailed"`
	Status          int    `json:"status"`
	Tokens          string `json:"tokens"`
	RewardsTokens   string `json:"rewards_tokens"`
	DelegatorShares string `json:"delegator_shares"`
	Description     struct {
		Moniker string `json:"moniker"`
	} `json:"description"`
	Commission struct {
		CommissionRates struct {
			Rate          string `json:"rate"`
			MaxRate       string `json:"max_rate"`
			MaxChangeRate string `json:"max_change_rate"`
		} `json:"commission_rates"`
		UpdateTime string `json:"update_time"`
	} `json:"commission"`
	SupportTokenType int `json:"support_token_type"`
}

type ValidatorResponse struct {
	Validator ValidatorInfo `json:"validator"`
}

type ValidatorsResponse struct {
	Validators []ValidatorInfo `json:"validators"`
	Pagination Pagination      `json:"pagination"`
}

type DelegationInfo struct {
	Delegation struct {
		DelegatorAddress string `json:"delegator_address"`
		ValidatorAddress string `json:"validator_address"`
		Shares           string `json:"shares"`
		RewardsShares    string `json:"rewards_shares"`
	} `json:"delegation"`
	Balance struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balance"`
}

type DelegationResponse struct {
	DelegationResponse DelegationInfo `json:"delegation_response"`
}

type DelegationsResponse struct {
	DelegationResponses []DelegationInfo `json:"delegation_responses"`
	Pagination          Pagination       `json:"pagination"`
}

type PeriodDelegationInfo struct {
	PeriodDelegation struct {
		DelegatorAddress   string `json:"delegator_address"`
		ValidatorAddress   string `json:"validator_address"`
		PeriodDelegationID string `json:"period_delegation_id"`
		PeriodType         int    `json:"period_type"`
		Shares             string `json:"shares"`
		RewardsShares      string `json:"rewards_shares"`
		EndTime            string `json:"end_time"`
	} `json:"period_delegation"`
	Balance struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balance"`
}

type PeriodDelegationResponse struct {
	PeriodDelegationResponse PeriodDelegationInfo `json:"period_delegation_response"`
}

type PeriodDelegationsResponse struct {
	PeriodDelegationResponses []PeriodDelegationInfo `json:"period_delegation_responses"`
	Pagination                Pagination             `json:"pagination"`
}

func ParsePaginationParams(c *gin.Context) map[string]string {
	params := make(map[string]string)

	if key := c.Query("pagination.key"); key != "" {
		params["pagination.key"] = key
	}
	if offset := c.Query("pagination.offset"); offset != "" {
		params["pagination.offset"] = offset
	}
	if limit := c.Query("pagination.limit"); limit != "" {
		params["pagination.limit"] = limit
	}
	if countTotal := c.Query("pagination.count_total"); countTotal != "" {
		params["pagination.count_total"] = countTotal
	}
	if reverse := c.Query("pagination.reverse"); reverse != "" {
		params["pagination.reverse"] = reverse
	}

	return params
}

type UnbondingDelegationsResponse struct {
	UnbondingResponses []struct {
		DelegatorAddress string `json:"delegator_address"`
		ValidatorAddress string `json:"validator_address"`
		Entries          []struct {
			CreationHeight string `json:"creation_height"`
			CompletionTime string `json:"completion_time"`
			InitialBalance string `json:"initial_balance"`
			Balance        string `json:"balance"`
			UnbondingID    string `json:"unbonding_id"`
		} `json:"entries"`
	} `json:"unbonding_responses"`
	Pagination Pagination `json:"pagination"`
}

func GetDistributionParams(apiEndpoint string) (*QueryResponse[DistributionParamsResponse], error) {
	resp, err := callAPI(apiEndpoint, "/distribution/params", nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[DistributionParamsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetMintParams(apiEndpoint string) (*QueryResponse[MintParamsResponse], error) {
	resp, err := callAPI(apiEndpoint, "/mint/params", nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[MintParamsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingParams(apiEndpoint string) (*QueryResponse[StakingParamsResponse], error) {
	resp, err := callAPI(apiEndpoint, "/staking/params", nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[StakingParamsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingPool(apiEndpoint string) (*QueryResponse[StakingPoolResponse], error) {
	resp, err := callAPI(apiEndpoint, "/staking/pool", nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[StakingPoolResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidators(apiEndpoint string, params map[string]string) (*QueryResponse[ValidatorsResponse], error) {
	resp, err := callAPI(apiEndpoint, "/staking/validators", params)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[ValidatorsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidator(apiEndpoint, validatorAddr string) (*QueryResponse[ValidatorResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s", validatorAddr), nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[ValidatorResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidatorDelegations(apiEndpoint, validatorAddr string, params map[string]string) (*QueryResponse[DelegationsResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegations", validatorAddr), params)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[DelegationsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingDelegation(apiEndpoint, validatorAddr, delegatorAddr string) (*QueryResponse[DelegationResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegations/%s", validatorAddr, delegatorAddr), nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[DelegationResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidatorDelegatorPeriodDelegations(apiEndpoint, validatorAddr, delegatorAddr string, params map[string]string) (*QueryResponse[PeriodDelegationsResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegators/%s/period_delegations", validatorAddr, delegatorAddr), params)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[PeriodDelegationsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidatorDelegatorPeriodDelegation(apiEndpoint, validatorAddr, delegatorAddr, delegationID string) (*QueryResponse[PeriodDelegationResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegators/%s/period_delegations/%s", validatorAddr, delegatorAddr, delegationID), nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[PeriodDelegationResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingDelegatorDelegations(apiEndpoint, delegatorAddr string, params map[string]string) (*QueryResponse[DelegationsResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/delegations/%s", delegatorAddr), params)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[DelegationsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingDelegatorUnbondingDelegations(apiEndpoint, delegatorAddr string, params map[string]string) (*QueryResponse[UnbondingDelegationsResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/delegators/%s/unbonding_delegations", delegatorAddr), params)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[UnbondingDelegationsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func callAPI(apiEndpoint, apiURL string, params map[string]string) (*http.Response, error) {
	reqURL, err := buildURL(apiEndpoint, apiURL, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metrics.RPCRequestErrorCounter.WithLabelValues(reqURL)
		return nil, err
	}

	if resp.StatusCode >= 400 {
		metrics.RPCRequestErrorCounter.WithLabelValues(reqURL)
	}

	return resp, nil
}

func buildURL(domain, path string, params map[string]string) (string, error) {
	baseURL, err := url.Parse(domain)
	if err != nil {
		return "", fmt.Errorf("invalid domain: %w", err)
	}

	baseURL.Path = path

	query := url.Values{}
	for key, value := range params {
		query.Add(key, value)
	}
	baseURL.RawQuery = query.Encode()

	return baseURL.String(), nil
}
