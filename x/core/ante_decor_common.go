package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/vmauth"
)

var (
	DefaultFees = sdk.Coins{sdk.NewCoin(config.MainDenom, sdk.NewInt(1))}
	// simulation signature values used to estimate gas consumption
	simSecp256k1Pubkey secp256k1.PubKeySecp256k1
	simSecp256k1Sig    [64]byte
)

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak vmauth.Keeper, addr sdk.AccAddress) (exported.Account, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}
	return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
}
