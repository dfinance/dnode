package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NewGovHandler creates proposal type handler for Gov module.
func NewGovHandler(keeper Keeper) gov.Handler {
	return func(ctx sdk.Context, c govTypes.Content) error {
		if c.ProposalRoute() != ModuleName {
			return fmt.Errorf("invalid proposal route %q for module %q", c.ProposalRoute(), ModuleName)
		}

		switch p := c.(type) {
		case StdlibUpdateProposal:
			return handleUpdateStdlibProposalDryRun(ctx, keeper, p)
		default:
			return fmt.Errorf("unsupported proposal content type %q for module %q", c.ProposalType(), ModuleName)
		}
	}
}

// handleUpdateStdlibProposalDryRun handles DVM stdlib update proposal: DVM validation and scheduling.
func handleUpdateStdlibProposalDryRun(ctx sdk.Context, keeper Keeper, p StdlibUpdateProposal) error {
	logger := keeper.Logger(ctx)

	if err := p.ValidateBasic(); err != nil {
		return sdkErrors.Wrapf(ErrGovInvalidProposal, "StdlibUpdateProposal validation: %v", err)
	}

	if err := keeper.ScheduleProposal(ctx, p); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("proposal scheduled:\n%s", p.String()))

	return nil
}
