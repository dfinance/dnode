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
	OperatedValidator *staking.Validator
	Delegations       []*staking.DelegationResponse
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

// AddDelegation appends delegation / updates an existing one.
func (a *SimAccount) AddDelegation(delegation *staking.DelegationResponse) {
	for i := 0; i < len(a.Delegations); i++ {
		existingDelegation := a.Delegations[i]
		if existingDelegation.ValidatorAddress.Equals(delegation.ValidatorAddress) {
			a.Delegations[i] = delegation
			return
		}
	}

	a.Delegations = append(a.Delegations, delegation)
}

func (a SimAccount) HasEnoughCoins(amount sdk.Coin) bool {
	accCoin := a.Coins.AmountOf(amount.Denom)
	if accCoin.LT(amount.Amount) {
		return false
	}
	return true
}
