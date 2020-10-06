package simulator

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	Day   = 24 * time.Hour
	Week  = 7 * Day
	Month = 4 * Week
	Year  = 12 * Month
)

// GetAllAccounts returns all known to Simulator accounts.
func (s *Simulator) GetAllAccounts() SimAccounts {
	return s.accounts
}

// GetAllValidators returns all known to Simulator validators.
func (s *Simulator) GetAllValidators() SimValidators {
	validators := make(SimValidators, 0)
	for _, acc := range s.accounts {
		if !acc.IsValOperator() {
			continue
		}
		validators = append(validators, acc.OperatedValidator)
	}

	return validators
}

// GetValidators returns known to Simulator validators filtered by status.
func (s *Simulator) GetValidators(bonded, unbonding, unbonded bool) SimValidators {
	validators := make(SimValidators, 0)
	for _, acc := range s.accounts {
		if acc.OperatedValidator != nil {
			add := false
			switch acc.OperatedValidator.GetStatus() {
			case sdk.Bonded:
				if bonded {
					add = true
				}
			case sdk.Unbonding:
				if unbonding {
					add = true
				}
			case sdk.Unbonded:
				if unbonded {
					add = true
				}
			}

			if add {
				validators = append(validators, acc.OperatedValidator)
			}
		}
	}

	return validators
}

// FormatCoin formats coin to decimal string.
func (s *Simulator) FormatCoin(coin sdk.Coin) string {
	return s.FormatIntDecimals(coin.Amount, s.stakingAmountDecimalsRatio) + coin.Denom
}

// FormatCoins formats coins to decimal string.
func (s *Simulator) FormatCoins(coins sdk.Coins) string {
	out := make([]string, 0, len(coins))
	for _, coin := range coins {
		out = append(out, s.FormatIntDecimals(coin.Amount, s.stakingAmountDecimalsRatio)+coin.Denom)
	}

	return strings.Join(out, ",")
}

// FormatIntDecimals converts sdk.Int to sdk.Dec using convert ratio and returns a string representation.
func (s *Simulator) FormatIntDecimals(value sdk.Int, decRatio sdk.Dec) string {
	valueDec := sdk.NewDecFromInt(value)
	fixedDec := valueDec.Mul(decRatio)

	return fixedDec.String()
}

func (s *Simulator) FormatDecDecimals(value sdk.Dec, decRatio sdk.Dec) string {
	fixedDec := value.Mul(decRatio)

	return fixedDec.String()
}
