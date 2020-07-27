package v0_7

import (
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

// Migrate migrates exported state from v0.6.X to a v0.7.0 genesis state.
func Migrate(appState genutil.AppMap) (genutil.AppMap, error) {
	//cdcOld := codec.New()
	//codec.RegisterCrypto(cdcOld)
	//
	//cdcNew := codec.New()
	//codec.RegisterCrypto(cdcNew)
	//
	//// migrate multisig module
	//moduleName := multisig.ModuleName
	//if stateOldBz := appState[moduleName]; stateOldBz != nil {
	//	var oldState v06multisig.GenesisState
	//	if err := cdcOld.UnmarshalJSON(stateOldBz, &oldState); err != nil {
	//		return nil, fmt.Errorf("module %q: oldState JSON unmarshal: %w", moduleName, err)
	//	}
	//
	//	newState, err := v07multisig.Migrate(oldState)
	//	if err != nil {
	//		return nil, fmt.Errorf("module %q: migration: %w", moduleName, err)
	//	}
	//
	//	stateNewBz, err := cdcNew.MarshalJSON(newState)
	//	if err != nil {
	//		return nil, fmt.Errorf("module %q: newState JSON marshal: %w", moduleName, err)
	//	}
	//
	//	delete(appState, moduleName)
	//	appState[moduleName] = stateNewBz
	//}

	return appState, nil
}
