package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/dfinance/dnode/x/vmauth"
)

// Consume parameter-defined amount of gas for each signature according to the passed-in SignatureVerificationGasConsumer function
// before calling the next AnteHandler
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type SigGasConsumeDecorator struct {
	ak             vmauth.Keeper
	sigGasConsumer ante.SignatureVerificationGasConsumer
}

func NewSigGasConsumeDecorator(ak vmauth.Keeper, sigGasConsumer ante.SignatureVerificationGasConsumer) SigGasConsumeDecorator {
	return SigGasConsumeDecorator{
		ak:             ak,
		sigGasConsumer: sigGasConsumer,
	}
}

func (sgcd SigGasConsumeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	sigTx, ok := tx.(ante.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	params := sgcd.ak.GetParams(ctx)
	sigs := sigTx.GetSignatures()

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signerAddrs := sigTx.GetSigners()

	for i, sig := range sigs {
		signerAcc, err := GetSignerAcc(ctx, sgcd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}
		pubKey := signerAcc.GetPubKey()

		if simulate && pubKey == nil {
			// In simulate mode the transaction comes with no signatures, thus if the
			// account's pubkey is nil, both signature verification and gasKVStore.Set()
			// shall consume the largest amount, i.e. it takes more gas to verify
			// secp256k1 keys than ed25519 ones.
			pubKey = simSecp256k1Pubkey
		}
		err = sgcd.sigGasConsumer(ctx.GasMeter(), sig, pubKey, params)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
