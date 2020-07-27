package v0_7

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/dfinance/dnode/x/fake"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/orders"
)

// Migrate migrates exported state from v0.6.X to a v0.7.0 genesis state.
func Migrate(appState genutil.AppMap) (genutil.AppMap, error) {
	cdcOld := codec.New()
	codec.RegisterCrypto(cdcOld)

	cdcNew := codec.New()
	codec.RegisterCrypto(cdcNew)

	// migrate multisig module
	moduleName := multisig.ModuleName
	if stateOldBz := appState[moduleName]; stateOldBz != nil {
		var oldState multisig.GenesisStateV06
		if err := cdcOld.UnmarshalJSON(stateOldBz, &oldState); err != nil {
			return nil, fmt.Errorf("module %q: oldState JSON unmarshal: %w", moduleName, err)
		}

		newState, err := multisig.MigrateV06ToV07(oldState)
		if err != nil {
			return nil, fmt.Errorf("module %q: migration: %w", moduleName, err)
		}

		stateNewBz, err := cdcNew.MarshalJSON(newState)
		if err != nil {
			return nil, fmt.Errorf("module %q: newState JSON marshal: %w", moduleName, err)
		}

		delete(appState, moduleName)
		appState[moduleName] = stateNewBz
	}

	// migrate orders module
	moduleName = orders.ModuleName
	if stateOldBz := appState[moduleName]; stateOldBz != nil {
		var oldState orders.GenesisStateV06
		if err := cdcOld.UnmarshalJSON(stateOldBz, &oldState); err != nil {
			return nil, fmt.Errorf("module %q: oldState JSON unmarshal: %w", moduleName, err)
		}

		newState, err := orders.MigrateV06ToV07(oldState)
		if err != nil {
			return nil, fmt.Errorf("module %q: migration: %w", moduleName, err)
		}

		stateNewBz, err := cdcNew.MarshalJSON(newState)
		if err != nil {
			return nil, fmt.Errorf("module %q: newState JSON marshal: %w", moduleName, err)
		}

		delete(appState, moduleName)
		appState[moduleName] = stateNewBz
	}

	moduleName = fake.ModuleName
	if stateOldBz := appState[moduleName]; stateOldBz == nil {
		newState := fake.DefaultGenesis()
		stateNewBz, err := cdcNew.MarshalJSON(newState)
		if err != nil {
			return nil, fmt.Errorf("module %q: newState JSON marshal: %w", moduleName, err)
		}

		appState[moduleName] = stateNewBz
	} else {
		return nil, fmt.Errorf("module %q: oldState not nil (should not exist in prev version)", moduleName)
	}

	return appState, nil
}
