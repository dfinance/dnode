package simulator

import (
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/libs/log"
)

type SimOption func(s *Simulator)

func InMemoryDBOption() SimOption {
	return func(s *Simulator) {
		s.useInMemDB = true
	}
}

func BlockTimeOption(min, max time.Duration) SimOption {
	return func(s *Simulator) {
		s.minBlockDur = min
		s.maxBlockDur = max
	}
}

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
				Name:  "account_" + strconv.Itoa(int(i+1)),
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

func MintParamsOption(modifier func(state *mint.GenesisState)) SimOption {
	return func(s *Simulator) {
		state := mint.GenesisState{}
		stateBz := s.genesisState[mint.ModuleName]
		s.cdc.MustUnmarshalJSON(stateBz, &state)

		modifier(&state)
		s.genesisState[mint.ModuleName] = s.cdc.MustMarshalJSON(state)
	}
}

func StakingParamsOption(modifier func(state *staking.GenesisState)) SimOption {
	return func(s *Simulator) {
		state := staking.GenesisState{}
		stateBz := s.genesisState[staking.ModuleName]
		s.cdc.MustUnmarshalJSON(stateBz, &state)

		modifier(&state)
		s.genesisState[staking.ModuleName] = s.cdc.MustMarshalJSON(state)
	}
}

func DistributionParamsOption(modifier func(state *distribution.GenesisState)) SimOption {
	return func(s *Simulator) {
		state := distribution.GenesisState{}
		stateBz := s.genesisState[distribution.ModuleName]
		s.cdc.MustUnmarshalJSON(stateBz, &state)

		modifier(&state)
		s.genesisState[distribution.ModuleName] = s.cdc.MustMarshalJSON(state)
	}
}
