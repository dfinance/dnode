package simulator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/tendermint/tendermint/libs/log"
)

type SimOption func(s *Simulator)

func InvariantCheckPeriodOption(period uint) SimOption {
	return func(s *Simulator) {
		s.invariantCheckPeriod = period
	}
}

func LogOption(option log.Option) SimOption {
	return func(s *Simulator) {
		s.logOptions = append(s.logOptions, option)
	}
}

func OperationsOption(ops ...*SimOperation) SimOption {
	return func(s *Simulator) {
		s.operations = append(s.operations, ops...)
	}
}

func GenerateWalletAccountsOption(walletsQuantity, poaValidatorsQuantity uint, genCoins sdk.Coins) SimOption {
	return func(s *Simulator) {
		for i := uint(0); i < walletsQuantity; i++ {
			acc := &SimAccount{
				Coins: genCoins,
			}
			if poaValidatorsQuantity > 0 {
				acc.IsPoAValidator = true
				poaValidatorsQuantity--
			}

			s.accounts = append(s.accounts, acc)
		}
	}
}

func NodeValidatorConfigOption(config SimValidatorConfig) SimOption {
	return func(s *Simulator) {
		s.nodeValidatorConfig = config
	}
}

func MintParamsOption(params mint.Params) SimOption {
	return func(s *Simulator) {
		state := mint.GenesisState{}
		stateBz := s.genesisState[mint.ModuleName]
		s.cdc.MustUnmarshalJSON(stateBz, &state)

		state.Params = params
		s.genesisState[mint.ModuleName] = s.cdc.MustMarshalJSON(state)
	}
}

func DistributionParamsOption(params distribution.Params) SimOption {
	return func(s *Simulator) {
		state := distribution.GenesisState{}
		stateBz := s.genesisState[distribution.ModuleName]
		s.cdc.MustUnmarshalJSON(stateBz, &state)

		state.Params = params
		s.genesisState[distribution.ModuleName] = s.cdc.MustMarshalJSON(state)
	}
}
