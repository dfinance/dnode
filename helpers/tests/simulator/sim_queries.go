package simulator

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (s *Simulator) RunQuery(requestData interface{}, path string, responseValue interface{}) abci.ResponseQuery {
	resp := s.app.Query(abci.RequestQuery{
		Data: codec.MustMarshalJSONIndent(s.cdc, requestData),
		Path: path,
	})

	if responseValue != nil && resp.IsOK() {
		require.NoError(s.t, s.cdc.UnmarshalJSON(resp.Value, responseValue))
	}

	return resp
}

// QueryAuthAccount queries account by address.
func (s *Simulator) QueryAuthAccount(addr sdk.AccAddress) (res auth.BaseAccount) {
	resp := s.RunQuery(
		auth.QueryAccountParams{
			Address: addr,
		},
		"/custom/"+auth.QuerierRoute+"/"+auth.QueryAccount,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryStakeValidators queries all validators with specified status and pagination.
func (s *Simulator) QueryStakeValidators(page, limit int, status string) []staking.Validator {
	res := make([]staking.Validator, 0)
	resp := s.RunQuery(
		staking.QueryValidatorsParams{
			Page:   page,
			Limit:  limit,
			Status: status,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryValidators,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryStakeValidator queries validator by its operator address.
func (s *Simulator) QueryStakeValidator(valAddr sdk.ValAddress) (res staking.Validator) {
	resp := s.RunQuery(
		staking.QueryValidatorParams{
			ValidatorAddr: valAddr,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryValidator,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryStakePools queries bonded / unbonded pool state.
func (s *Simulator) QueryStakePools() (res staking.Pool) {
	resp := s.RunQuery(
		nil,
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryPool,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryStakeDelegation queries delegation for specified delegator and validator.
func (s *Simulator) QueryStakeDelegation(simAcc *SimAccount, val *staking.Validator) (res staking.DelegationResponse) {
	require.NotNil(s.t, simAcc)
	require.NotNil(s.t, val)

	resp := s.RunQuery(
		staking.QueryBondsParams{
			DelegatorAddr: simAcc.Address,
			ValidatorAddr: val.OperatorAddress,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryDelegation,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryStakeDelDelegations queries delegator delegations.
func (s *Simulator) QueryStakeDelDelegations(delegator sdk.AccAddress) (res staking.DelegationResponses) {
	resp := s.RunQuery(
		staking.QueryDelegatorParams{
			DelegatorAddr: delegator,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryDelegatorDelegations,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return
}

// QueryStakeValDelegations queries delegations for specified validator.
func (s *Simulator) QueryStakeValDelegations(val *staking.Validator) (res staking.DelegationResponses) {
	require.NotNil(s.t, val)

	resp := s.RunQuery(
		staking.QueryValidatorParams{
			ValidatorAddr: val.OperatorAddress,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryValidatorDelegations,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryStakeRedelegations queries redelegations.
func (s *Simulator) QueryStakeRedelegations(delegator sdk.AccAddress, valSrc, valDst sdk.ValAddress) (res staking.RedelegationResponses) {
	resp := s.RunQuery(
		staking.QueryRedelegationParams{
			DelegatorAddr:    delegator,
			SrcValidatorAddr: valSrc,
			DstValidatorAddr: valDst,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryRedelegations,
		&res,
	)

	if resp.Code == staking.ErrNoRedelegation.ABCICode() {
		return
	}
	require.True(s.t, resp.IsOK())

	return
}

// QueryStakeDelUnbondingDelegations queries delegator unbonding delegations.
func (s *Simulator) QueryStakeDelUnbondingDelegations(delegatorAddr sdk.AccAddress) (res staking.UnbondingDelegations) {
	resp := s.RunQuery(
		staking.QueryDelegatorParams{
			DelegatorAddr: delegatorAddr,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryDelegatorUnbondingDelegations,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryStakeDelHasUnbondingDelegation check if delegator has unbonding delegation for specified validator.
func (s *Simulator) QueryStakeDelHasUnbondingDelegation(delAddr sdk.AccAddress, valAddr sdk.ValAddress) bool {
	res := staking.QueryBondsParams{}
	resp := s.RunQuery(
		staking.QueryBondsParams{
			DelegatorAddr: delAddr,
			ValidatorAddr: valAddr,
		},
		"/custom/"+staking.QuerierRoute+"/"+staking.QueryUnbondingDelegation,
		&res,
	)

	if resp.Code == staking.ErrNoUnbondingDelegation.ABCICode() {
		return false
	}
	require.True(s.t, resp.IsOK())

	return true
}

// QueryMintParams queries mint parameters.
func (s *Simulator) QueryMintParams() (res mint.Params) {
	resp := s.RunQuery(
		nil,
		"/custom/"+mint.QuerierRoute+"/"+mint.QueryParameters,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryMintParams queries mint annual provisions.
func (s *Simulator) QueryMintAnnualProvisions() (res sdk.Dec) {
	resp := s.RunQuery(
		nil,
		"/custom/"+mint.QuerierRoute+"/"+mint.QueryAnnualProvisions,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryMintParams queries mint blocksPerYear estimation.
func (s *Simulator) QueryMintBlocksPerYearEstimation() (res uint64) {
	resp := s.RunQuery(
		nil,
		"/custom/"+mint.QuerierRoute+"/"+mint.QueryBlocksPerYear,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryDistPool queries distribution pools supply.
func (s *Simulator) QueryDistPool(poolName distribution.RewardPoolName) (res sdk.DecCoins) {
	resp := s.RunQuery(
		nil,
		"/custom/"+distribution.QuerierRoute+"/"+distribution.QueryPool+"/"+poolName.String(),
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryDistDelReward queries current delegator rewards for specified validator.
func (s *Simulator) QueryDistDelReward(accAddr sdk.AccAddress, valAddr sdk.ValAddress) (res sdk.DecCoins) {
	resp := s.RunQuery(
		distribution.QueryDelegationRewardsParams{
			DelegatorAddress: accAddr,
			ValidatorAddress: valAddr,
		},
		"/custom/"+distribution.QuerierRoute+"/"+distribution.QueryDelegationRewards,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return
}

// QueryDistDelRewards queries current delegator rewards.
func (s *Simulator) QueryDistDelRewards(acc sdk.AccAddress) (res distribution.QueryDelegatorTotalRewardsResponse) {
	resp := s.RunQuery(
		distribution.QueryDelegatorParams{
			DelegatorAddress: acc,
		},
		"/custom/"+distribution.QuerierRoute+"/"+distribution.QueryDelegatorTotalRewards,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return
}

// QueryDistrValCommission queries current validator commission rewards.
func (s *Simulator) QueryDistrValCommission(val sdk.ValAddress) (res distribution.ValidatorAccumulatedCommission) {
	resp := s.RunQuery(
		distribution.QueryValidatorCommissionParams{
			ValidatorAddress: val,
		},
		"/custom/"+distribution.QuerierRoute+"/"+distribution.QueryValidatorCommission,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}

// QueryDistPool queries supply total supply.
func (s *Simulator) QuerySupplyTotal() (res sdk.Coins) {
	resp := s.RunQuery(
		supply.QueryTotalSupplyParams{
			Page:  1,
			Limit: 50,
		},
		"/custom/"+supply.QuerierRoute+"/"+supply.QueryTotalSupply,
		&res,
	)
	require.True(s.t, resp.IsOK())

	return res
}
