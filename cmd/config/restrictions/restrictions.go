package restrictions

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
	"github.com/dfinance/dnode/x/currencies"
)

// Custom restriction params for application
type AppRestrictions struct {
	CustomMsgVerifiers func(msg sdk.Msg) error
	MsgDeniedList      map[string][]string
	ParamsProposal     params.RestrictedParams
	DisabledTxCmd      []string
	DisabledQueryCmd   []string
}

// GetEmptyAppRestriction returns AppRestrictions with no restrictions.
func GetEmptyAppRestriction() AppRestrictions {
	return AppRestrictions{
		DisabledTxCmd:      []string{},
		DisabledQueryCmd:   []string{},
		MsgDeniedList:      map[string][]string{},
		ParamsProposal:     params.RestrictedParams{},
		CustomMsgVerifiers: func(msg sdk.Msg) error { return nil },
	}
}

//GetAppRestrictions returns predefined parameter for remove or restrict standard app parameters.
func GetAppRestrictions() AppRestrictions {
	return AppRestrictions{
		DisabledTxCmd: []string{
			distribution.ModuleName,
		},
		DisabledQueryCmd: []string{},
		MsgDeniedList: map[string][]string{
			distribution.ModuleName: {
				distribution.MsgWithdrawDelegatorReward{}.Type(),
				distribution.MsgWithdrawValidatorCommission{}.Type(),
				distribution.TypeMsgFundPublicTreasuryPool,
				distribution.MsgSetWithdrawAddress{}.Type(),
			},
			currencies.ModuleName: {
				currencies.MsgWithdrawCurrency{}.Type(),
			},
		},
		ParamsProposal: params.RestrictedParams{
			params.RestrictedParam{Subspace: distribution.ModuleName, Key: string(distribution.ParamKeyValidatorsPoolTax)},
			params.RestrictedParam{Subspace: distribution.ModuleName, Key: string(distribution.ParamKeyLiquidityProvidersPoolTax)},
			params.RestrictedParam{Subspace: distribution.ModuleName, Key: string(distribution.ParamKeyPublicTreasuryPoolTax)},
			params.RestrictedParam{Subspace: distribution.ModuleName, Key: string(distribution.ParamKeyHARPTax)},
			params.RestrictedParam{Subspace: distribution.ModuleName, Key: string(distribution.ParamKeyFoundationNominees)},
			params.RestrictedParam{Subspace: mint.ModuleName, Key: string(mint.KeyFoundationAllocationRatio)},
			params.RestrictedParam{Subspace: mint.ModuleName, Key: string(mint.KeyStakingTotalSupplyShift)},
		},
		CustomMsgVerifiers: func(msg sdk.Msg) error {
			switch msg := msg.(type) {
			case bank.MsgSend:
				for i := range msg.Amount {
					if msg.Amount.GetDenomByIndex(i) == defaults.StakingDenom || msg.Amount.GetDenomByIndex(i) == defaults.LiquidityProviderDenom {
						return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "bank transactions are disallowed for %s token", msg.Amount.GetDenomByIndex(i))
					}
				}
			case gov.MsgDeposit:
				for i := range msg.Amount {
					if msg.Amount.GetDenomByIndex(i) != defaults.StakingDenom {
						return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "gov deposit only allowed for %s token", defaults.StakingDenom)
					}
				}
			}

			return nil
		},
	}
}
