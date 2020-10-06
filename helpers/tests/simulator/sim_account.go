package simulator

import (
	"math/rand"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
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
	OperatedValidator *SimValidator
	Delegations       []staking.DelegationResponse
}

// IsValOperator checks if account is a validator operator.
func (a *SimAccount) IsValOperator() bool {
	return a.OperatedValidator != nil
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

// GetSortedDelegations returns delegations sorted by tokens balance list.
func (a *SimAccount) GetSortedDelegations(bondingTokens, desc bool) staking.DelegationResponses {
	tmpDels := make(staking.DelegationResponses, len(a.Delegations))
	copy(tmpDels, a.Delegations)

	sort.Slice(tmpDels, func(i, j int) bool {
		if bondingTokens {
			if tmpDels[i].BondingBalance.Amount.GT(tmpDels[j].BondingBalance.Amount) {
				return desc
			}
			return !desc
		}

		if tmpDels[i].LPBalance.Amount.GT(tmpDels[j].LPBalance.Amount) {
			return desc
		}
		return !desc
	})

	return tmpDels
}

// GetShuffledDelegations returns shuffled delegations list.
func (a *SimAccount) GetShuffledDelegations(bondingTokens bool) staking.DelegationResponses {
	tmpDels := make(staking.DelegationResponses, len(a.Delegations))
	copy(tmpDels, a.Delegations)

	for i := range tmpDels {
		j := rand.Intn(i + 1)
		tmpDels[i], tmpDels[j] = tmpDels[j], tmpDels[i]
	}

	return tmpDels
}

type SimAccounts []*SimAccount

// GetByAddress returns account by address.
func (a SimAccounts) GetByAddress(address sdk.AccAddress) *SimAccount {
	for _, acc := range a {
		if acc.Address.Equals(address) {
			return acc
		}
	}

	return nil
}

// GetRandom returns randomly selected account.
func (a SimAccounts) GetRandom() *SimAccount {
	aMaxIndex := len(a) - 1

	return a[rand.Intn(aMaxIndex)]
}

// GetShuffled returns random sorted accounts list.
func (a SimAccounts) GetShuffled() SimAccounts {
	tmpAcc := make(SimAccounts, len(a))
	copy(tmpAcc, a)

	for i := range tmpAcc {
		j := rand.Intn(i + 1)
		tmpAcc[i], tmpAcc[j] = tmpAcc[j], tmpAcc[i]
	}

	return tmpAcc
}

// GetAccountsSortedByBalance returns account sorted by staking denom list.
func (a SimAccounts) GetSortedByBalance(denom string, desc bool) SimAccounts {
	tmpAccs := make(SimAccounts, len(a))
	copy(tmpAccs, a)

	sort.Slice(tmpAccs, func(i, j int) bool {
		iBalance := tmpAccs[i].Coins.AmountOf(denom)
		jBalance := tmpAccs[j].Coins.AmountOf(denom)

		if iBalance.GT(jBalance) {
			return desc
		}
		return !desc
	})

	return tmpAccs
}

// UpdateAccount updates account balance and active delegations.
func (s *Simulator) UpdateAccount(simAcc *SimAccount) {
	require.NotNil(s.t, simAcc)

	updAcc := s.QueryAuthAccount(simAcc.Address)
	simAcc.Coins = updAcc.GetCoins()
	simAcc.Delegations = s.QueryStakeDelDelegations(simAcc.Address)
}
