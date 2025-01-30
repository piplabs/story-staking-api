package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gin-gonic/gin"

	minttypes "github.com/piplabs/story/client/x/mint/types"
)

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

type QueryResponse[T any] struct {
	Code  int    `json:"code"`
	Msg   T      `json:"msg"`
	Error string `json:"error"`
}

func GetDistributionParams(apiEndpoint string) (*QueryResponse[distributiontypes.QueryParamsResponse], error) {
	resp, err := callAPI(apiEndpoint, "/distribution/params", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[distributiontypes.QueryParamsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetMintParams(apiEndpoint string) (*QueryResponse[minttypes.QueryParamsResponse], error) {
	resp, err := callAPI(apiEndpoint, "/mint/params", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[minttypes.QueryParamsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingPool(apiEndpoint string) (*QueryResponse[stakingtypes.QueryPoolResponse], error) {
	resp, err := callAPI(apiEndpoint, "/staking/pool", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.QueryPoolResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidators(apiEndpoint string, params map[string]string) (*QueryResponse[stakingtypes.QueryValidatorsResponse], error) {
	resp, err := callAPI(apiEndpoint, "/staking/validators", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.QueryValidatorsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidator(apiEndpoint, validatorAddr string) (*QueryResponse[stakingtypes.QueryValidatorResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s", validatorAddr), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.QueryValidatorResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidatorDelegations(apiEndpoint, validatorAddr string, params map[string]string) (*QueryResponse[stakingtypes.QueryValidatorDelegationsResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegations", validatorAddr), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.QueryValidatorDelegationsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingDelegation(apiEndpoint, validatorAddr, delegatorAddr string) (*QueryResponse[stakingtypes.QueryDelegationResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegations/%s", validatorAddr, delegatorAddr), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.QueryDelegationResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidatorDelegatorPeriodDelegations(apiEndpoint, validatorAddr, delegatorAddr string, params map[string]string) (*QueryResponse[[]stakingtypes.PeriodDelegation], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegators/%s/period_delegations", validatorAddr, delegatorAddr), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[[]stakingtypes.PeriodDelegation] // TODO: update type
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingValidatorDelegatorPeriodDelegation(apiEndpoint, validatorAddr, delegatorAddr, delegationID string) (*QueryResponse[stakingtypes.PeriodDelegation], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/validators/%s/delegators/%s/period_delegations/%s", validatorAddr, delegatorAddr, delegationID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.PeriodDelegation] // TODO: update type
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingDelegatorDelegations(apiEndpoint, delegatorAddr string, params map[string]string) (*QueryResponse[stakingtypes.QueryDelegatorDelegationsResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/delegations/%s", delegatorAddr), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.QueryDelegatorDelegationsResponse]
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, err
	}

	if res.Code != http.StatusOK {
		return nil, errors.New(res.Error)
	}

	return &res, nil
}

func GetStakingDelegatorUnbondingDelegations(apiEndpoint, delegatorAddr string, params map[string]string) (*QueryResponse[stakingtypes.QueryDelegatorUnbondingDelegationsResponse], error) {
	resp, err := callAPI(apiEndpoint, fmt.Sprintf("/staking/delegators/%s/unbonding_delegations", delegatorAddr), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryResponse[stakingtypes.QueryDelegatorUnbondingDelegationsResponse]
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

	return http.DefaultClient.Do(req)
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
