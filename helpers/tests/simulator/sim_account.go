package simulator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type SimAccount struct {
	Name              string
	Address           sdk.AccAddress
	Number            uint64
	PrivateKey        secp256k1.PrivKeySecp256k1
	PublicKey         crypto.PubKey
	Coins             sdk.Coins
	IsPoAValidator    bool
	CreateValidator   bool
	OperatedValidator *staking.Validator
	Delegations       []staking.DelegationResponse
}

// HasDelegation checks if account has already delegated to the specified validator.
func (a *SimAccount) HasDelegation(valAddress sdk.ValAddress) bool {
	for _, del := range a.Delegations {
		if del.ValidatorAddress.Equals(valAddress) {
			return true
		}
	}

	return false
}
