package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/dfinance/dnode/x/common_vm"
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
func handleUpdateStdlibProposalDryRun(ctx sdk.Context, keeper Keeper, proposal StdlibUpdateProposal) error {
	logger := keeper.Logger(ctx)

	// DVM check (dry-run deploy)
	msg, err := getStdlibUpdateMsg(proposal)
	if err != nil {
		return err
	}
	if err := keeper.DeployContractDryRun(ctx, msg); err != nil {
		return fmt.Errorf("contract dry run deploy failed: %w", err)
	}

	// add proposal to queue
	if err := keeper.ScheduleProposal(ctx, proposal); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("proposal scheduled:\n%s", proposal.String()))

	return nil
}

// getStdlibUpdateMsg returns deploy message for stdlib update.
func getStdlibUpdateMsg(proposal StdlibUpdateProposal) (MsgDeployModule, error) {
	msg := NewMsgDeployModule(common_vm.Bech32ToLibra(common_vm.StdLibAddress), proposal.Code)
	if err := msg.ValidateBasic(); err != nil {
		return MsgDeployModule{}, fmt.Errorf("deploy message validation failed: %w", err)
	}

	return msg, nil
}
