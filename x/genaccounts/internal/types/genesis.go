package types

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
)

// State to Unmarshal
type GenesisState GenesisAccounts

// Get the genesis state from the expected app state.
func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	genesisState := GenesisState{}
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

// Set the genesis state within the expected app state.
func SetGenesisStateInAppState(cdc *codec.Codec, appState map[string]json.RawMessage, genesisState GenesisState) map[string]json.RawMessage {
	genesisStateBz := cdc.MustMarshalJSON(genesisState)
	appState[ModuleName] = genesisStateBz

	return appState
}

// Sanitize sorts accounts and coin sets.
func (gs GenesisState) Sanitize() {
	sort.Slice(gs, func(i, j int) bool {
		return gs[i].BaseAccount.AccountNumber < gs[j].BaseAccount.AccountNumber
	})

	for _, acc := range gs {
		acc.BaseAccount.Coins = acc.BaseAccount.Coins.Sort()
	}
}

// ValidateGenesis performs validation of genesis accounts, ensures that there are no duplicate accounts.
func ValidateGenesis(genesisState GenesisState) error {
	addrMap := make(map[string]bool, len(genesisState))
	for _, acc := range genesisState {
		addrStr := acc.BaseAccount.Address.String()

		// disallow any duplicate accounts
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}

		addrMap[addrStr] = true
	}
	return nil
}
