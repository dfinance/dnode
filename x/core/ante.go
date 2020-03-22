// Implements custom AnteHandler.
package core

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/vmauth"
)

var (
	DefaultFees     = sdk.Coins{sdk.NewCoin(config.MainDenom, sdk.NewInt(1))}
	simSecp256k1Sig [64]byte
)

// Custom antehandler catches and prevents transactions without fees and fees not in "dfi" currency
// After execution of custom logic, call standard auth.AnteHandler.
func NewAnteHandler(ak vmauth.VMAccountKeeper, supplyKeeper types.SupplyKeeper, sigGasConsumer auth.SignatureVerificationGasConsumer) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
		stdTx, ok := tx.(auth.StdTx)

		if !ok {
			newCtx = auth.SetGasMeter(simulate, ctx, 0)
			return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		// ignore genesis block.
		if ctx.BlockHeight() > 0 {
			if stdTx.Fee.Amount.IsZero() {
				newCtx = auth.SetGasMeter(simulate, ctx, 0)
				return newCtx, ErrFeeRequired().Result(), true
			}

			if !stdTx.Fee.Amount.DenomsSubsetOf(DefaultFees) {
				newCtx = auth.SetGasMeter(simulate, ctx, 0)
				return newCtx, ErrWrongFeeDenom(config.MainDenom).Result(), true
			}
		}

		return AuthAnteHandler(ak, supplyKeeper, sigGasConsumer)(ctx, tx, simulate)
	}
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func AuthAnteHandler(ak vmauth.VMAccountKeeper, supplyKeeper types.SupplyKeeper, sigGasConsumer auth.SignatureVerificationGasConsumer) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		if addr := supplyKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
			helpers.CrashWithMessage("%s module account has not been set", types.FeeCollectorName)
		}

		// all transactions must be of type auth.StdTx
		stdTx, ok := tx.(types.StdTx)
		if !ok {
			// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
			// during runTx.
			newCtx = auth.SetGasMeter(simulate, ctx, 0)
			return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		params := ak.GetParams(ctx)

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only ran on check tx.
		if ctx.IsCheckTx() && !simulate {
			res := auth.EnsureSufficientMempoolFees(ctx, stdTx.Fee)
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		newCtx = auth.SetGasMeter(simulate, ctx, stdTx.Fee.Gas)

		// AnteHandlers must have their own defer/recover in order for the BaseApp
		// to know how much gas was used! This is because the GasMeter is created in
		// the AnteHandler, but if it panics the context won't be set properly in
		// runTx's recover call.
		defer func() {
			if r := recover(); r != nil {
				switch rType := r.(type) {
				case sdk.ErrorOutOfGas:
					log := fmt.Sprintf(
						"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
						rType.Descriptor, stdTx.Fee.Gas, newCtx.GasMeter().GasConsumed(),
					)
					res = sdk.ErrOutOfGas(log).Result()

					res.GasWanted = stdTx.Fee.Gas
					res.GasUsed = newCtx.GasMeter().GasConsumed()
					abort = true
				default:
					helpers.CrashWithObject(r)
				}
			}
		}()

		if res := auth.ValidateSigCount(stdTx, params); !res.IsOK() {
			return newCtx, res, true
		}

		if err := tx.ValidateBasic(); err != nil {
			return newCtx, err.Result(), true
		}

		newCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(newCtx.TxBytes())), "txSize")

		if res := auth.ValidateMemo(stdTx, params); !res.IsOK() {
			return newCtx, res, true
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()
		signerAccs := make([]auth.Account, len(signerAddrs))
		isGenesis := ctx.BlockHeight() == 0

		// fetch first signer, who's going to pay the fees
		signerAccs[0], res = vmauth.GetSignerAcc(newCtx, ak, signerAddrs[0])
		if !res.IsOK() {
			return newCtx, res, true
		}

		// deduct the fees
		if !stdTx.Fee.Amount.IsZero() {
			res = auth.DeductFees(supplyKeeper, newCtx, signerAccs[0], stdTx.Fee.Amount)
			if !res.IsOK() {
				return newCtx, res, true
			}

			// reload the account as fees have been deducted
			signerAccs[0] = ak.GetAccount(newCtx, signerAccs[0].GetAddress())
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		stdSigs := stdTx.GetSignatures()

		for i := 0; i < len(stdSigs); i++ {
			// skip the fee payer, account is cached and fees were deducted already
			if i != 0 {
				signerAccs[i], res = vmauth.GetSignerAcc(newCtx, ak, signerAddrs[i])
				if !res.IsOK() {
					return newCtx, res, true
				}
			}

			// check signature, return account with incremented nonce
			signBytes := auth.GetSignBytes(newCtx.ChainID(), stdTx, signerAccs[i], isGenesis)
			signerAccs[i], res = processSig(newCtx, signerAccs[i], stdSigs[i], signBytes, simulate, params, sigGasConsumer)
			if !res.IsOK() {
				return newCtx, res, true
			}

			ak.SetAccount(newCtx, signerAccs[i])
		}

		// TODO: tx tags (?)
		return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false // continue...
	}
}

// verify the signature and increment the sequence. If the account doesn't have
// a pubkey, set it.
func processSig(
	ctx sdk.Context, acc auth.Account, sig auth.StdSignature, signBytes []byte, simulate bool, params auth.Params,
	sigGasConsumer auth.SignatureVerificationGasConsumer,
) (updatedAcc auth.Account, res sdk.Result) {

	pubKey, res := auth.ProcessPubKey(acc, sig, simulate)
	if !res.IsOK() {
		return nil, res
	}

	if err := acc.SetPubKey(pubKey); err != nil {
		return nil, sdk.ErrInternal("setting PubKey on signer's account").Result()
	}

	if simulate {
		// Simulated txs should not contain a signature and are not required to
		// contain a pubkey, so we must account for tx size of including a
		// StdSignature (Amino encoding) and simulate gas consumption
		// (assuming a SECP256k1 simulation key).
		consumeSimSigGas(ctx.GasMeter(), pubKey, sig, params)
	}

	if res := sigGasConsumer(ctx.GasMeter(), sig.Signature, pubKey, params); !res.IsOK() {
		return nil, res
	}

	if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed; verify correct account sequence and chain-id").Result()
	}

	if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
		helpers.CrashWithError(err)
	}

	return acc, res
}

func consumeSimSigGas(gasmeter sdk.GasMeter, pubkey crypto.PubKey, sig auth.StdSignature, params auth.Params) {
	simSig := auth.StdSignature{PubKey: pubkey}
	if len(sig.Signature) == 0 {
		simSig.Signature = simSecp256k1Sig[:]
	}

	sigBz := auth.ModuleCdc.MustMarshalBinaryLengthPrefixed(simSig)
	cost := sdk.Gas(len(sigBz) + 6)

	// If the pubkey is a multi-signature pubkey, then we estimate for the maximum
	// number of signers.
	if _, ok := pubkey.(multisig.PubKeyMultisigThreshold); ok {
		cost *= params.TxSigLimit
	}

	gasmeter.ConsumeGas(params.TxSizeCostPerByte*cost, "txSize")
}
