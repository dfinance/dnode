package v1_0

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	v03902distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_39-0_2"
	v03910distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_39-1_0"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	v03902staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_39-0_2"
	v03910staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_39-1_0"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
)

// Migrate migrates exported genesis state from Dfinance v0.7 Testnet to v1.0 Mainnet.
func Migrate(appState genutil.AppMap) (genutil.AppMap, error) {
	cdcOld := codec.New()
	codec.RegisterCrypto(cdcOld)

	cdcNew := codec.New()
	codec.RegisterCrypto(cdcNew)

	// Cosmos SDK modules
	// staking
	{
		moduleName := staking.ModuleName
		if stateOldBz := appState[moduleName]; stateOldBz != nil {
			migrationOpts := v03910staking.MigrateOptions{
				ParamsMaxSelfDelegationLvl: defaults.MaxSelfDelegationCoin.Amount,
			}

			var oldState v03902staking.GenesisState
			if err := cdcOld.UnmarshalJSON(stateOldBz, &oldState); err != nil {
				return nil, fmt.Errorf("module %q: oldState JSON unmarshal: %w", moduleName, err)
			}

			newState := v03910staking.Migrate(oldState, migrationOpts)
			stateNewBz, err := cdcNew.MarshalJSON(newState)
			if err != nil {
				return nil, fmt.Errorf("module %q: newState JSON marshal: %w", moduleName, err)
			}

			delete(appState, moduleName)
			appState[moduleName] = stateNewBz
		}
	}
	// distribution
	{
		moduleName := distribution.ModuleName
		if stateOldBz := appState[moduleName]; stateOldBz != nil {
			var oldState v03902distribution.GenesisState
			if err := cdcOld.UnmarshalJSON(stateOldBz, &oldState); err != nil {
				return nil, fmt.Errorf("module %q: oldState JSON unmarshal: %w", moduleName, err)
			}

			newState := v03910distribution.Migrate(oldState)
			stateNewBz, err := cdcNew.MarshalJSON(newState)
			if err != nil {
				return nil, fmt.Errorf("module %q: newState JSON marshal: %w", moduleName, err)
			}

			delete(appState, moduleName)
			appState[moduleName] = stateNewBz
		}
	}

	return appState, nil
}
