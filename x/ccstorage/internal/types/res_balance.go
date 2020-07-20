package types

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/lcs"
)

// ResBalance is a DVM resource, containing an account coins balance.
type ResBalance struct {
	Value *big.Int `json:"value" yaml:"balance"`
}

// Bytes returns balance lcs marshalled.
func (b ResBalance) Bytes() ([]byte, error) {
	bytes, err := lcs.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("lsc marshal: %w", err)
	}

	return bytes, nil
}

// NewResBalance unmarshals lcs representation to balance resource.
func NewResBalance(bz []byte) (ResBalance, error) {
	balance := ResBalance{}
	err := lcs.Unmarshal(bz, &balance)
	if err != nil {
		return ResBalance{}, fmt.Errorf("lsc unmarshal: %w", err)
	}

	return balance, nil
}

// Balance is a wrapped ResBalance with extra meta-data.
type Balance struct {
	Denom      string
	AccessPath *vm_grpc.VMAccessPath
	Resource   ResBalance
}

// Coin converts balance to coin.
func (b Balance) Coin() sdk.Coin {
	return sdk.NewCoin(b.Denom, sdk.NewIntFromBigInt(b.Resource.Value))
}

// ResourceBytes returns byte resource representation.
func (b Balance) ResourceBytes() ([]byte, error) {
	bz, err := b.Resource.Bytes()
	if err != nil {
		return nil, fmt.Errorf("balance %q: %w", b.Denom, err)
	}

	return bz, nil
}

// NewBalance creates a new balance object.
func NewBalance(denom string, accessPath *vm_grpc.VMAccessPath, resBytes []byte) (Balance, error) {
	if denom == "" {
		return Balance{}, fmt.Errorf("resouce for denom %q: denom is empty", denom)
	}
	if accessPath == nil {
		return Balance{}, fmt.Errorf("resouce for denom %q: accessPath is nil", denom)
	}

	res, err := NewResBalance(resBytes)
	if err != nil {
		return Balance{}, fmt.Errorf("resouce for denom %q: %w", denom, err)
	}

	return Balance{
		Denom:      denom,
		AccessPath: accessPath,
		Resource:   res,
	}, nil
}

// Balances is a slice type for balance.
type Balances []Balance

// Coins converts balances to sdk.Coins.
func (l Balances) Coins() sdk.Coins {
	coins := make(sdk.Coins, 0, len(l))

	// ignore zero values
	for _, balance := range l {
		if balance.Resource.Value.Cmp(sdk.ZeroInt().BigInt()) != 0 {
			coins = append(coins, balance.Coin())
		}
	}

	return coins
}
