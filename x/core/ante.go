// Implements custom AnteHandler.
package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/dfinance/dnode/x/vmauth"
)

// NewAnteHandler return custom AnteHandler.
// Adds DenomDecorator and uses standard decorators (standard AnteHandler).
// Some decorators are a copy of 'github.com/cosmos/cosmos-sdk/x/auth/ante' decorators, but using vmauth.VMAccountKeeper.
func NewAnteHandler(ak *vmauth.VMAccountKeeper, supplyKeeper types.SupplyKeeper, sigGasConsumer auth.SignatureVerificationGasConsumer) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		NewDenomDecorator(),
		ante.NewSetUpContextDecorator(),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewValidateMemoDecorator(*ak.AccountKeeper),      // as is: only uses ak.GetParams()
		NewConsumeGasForTxSizeDecorator(*ak),                  // copy: uses ak.GetAccount()
		NewSetPubKeyDecorator(*ak),                            // copy: uses ak.GetAccount()
		ante.NewValidateSigCountDecorator(*ak.AccountKeeper),  // as is: only uses ak.GetParams()
		NewDeductFeeDecorator(*ak, supplyKeeper),              // copy: uses ak.GetAccount()
		NewSigGasConsumeDecorator(*ak, sigGasConsumer),        // copy: uses ak.GetAccount()
		NewSigVerificationDecorator(*ak),                      // copy: uses ak.GetAccount()
		NewIncrementSequenceDecorator(*ak),                    // copy: uses ak.GetAccount(), ak.SetAccount()
	)
}
