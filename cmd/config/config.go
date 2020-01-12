// Configuration for WBD and WBCli.
package config

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	MainDenom            = "wings"
	MainPrefix           = "wallets"                                                                 // Main prefix for all addresses.
	Bech32PrefixAccAddr  = MainPrefix                                                                // Bech32 prefix for account addresses.
	Bech32PrefixAccPub   = MainPrefix + sdk.PrefixPublic                                             // Bech32 prefix for accounts pub keys.
	Bech32PrefixValAddr  = MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator                     // Bech32 prefix for validators addresses.
	Bech32PrefixValPub   = MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic  // Bech32 prefix for validator pub keys.
	Bech32PrefixConsAddr = MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus                    // Bech32 prefix for consensus addresses.
	Bech32PrefixConsPub  = MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic // Bech32 prefix for consensus pub keys.
)

// Initializing WB custom prefixes.
func InitBechPrefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
}
