package simulator

import "github.com/cosmos/cosmos-sdk/x/staking"

type SimValidatorConfig struct {
	Commission staking.CommissionRates
}
