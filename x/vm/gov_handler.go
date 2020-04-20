package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// New governance proposal handler for Gov module.
func NewGovHandler(keeper Keeper) gov.Handler {
	return func(ctx sdk.Context, c govTypes.Content) error {
		if c.ProposalRoute() != ModuleName {
			return fmt.Errorf("invalid proposal route %q for module %q", c.ProposalRoute(), ModuleName)
		}

		switch p := c.(type) {
		case ModuleUpdateProposal:
			return handleUpdateModuleProposalDryRun(ctx, keeper, p)
		case TestProposal:
			return handleTestProposalDryRun(ctx, keeper, p)
		default:
			return fmt.Errorf("unsupported proposal content type %q for module %q", c.ProposalType(), ModuleName)
		}
	}
}

func handleUpdateModuleProposalDryRun(ctx sdk.Context, keeper Keeper, p types.ModuleUpdateProposal) error {
	logger := keeper.Logger(ctx)

	pProposal := types.NewPlannedProposal(p, ProposalData{}, p.Plan)
	if err := pProposal.ValidateBasic(); err != nil {
		return err
	}

	if err := keeper.ScheduleProposal(ctx, pProposal); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("proposal scheduled: %s", pProposal.String()))

	return nil
}

func handleTestProposalDryRun(ctx sdk.Context, keeper Keeper, p types.TestProposal) error {
	logger := keeper.Logger(ctx)

	pProposal := types.NewPlannedProposal(p, ProposalData{}, p.Plan)
	if err := pProposal.ValidateBasic(); err != nil {
		return err
	}

	if err := keeper.ScheduleProposal(ctx, pProposal); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("proposal scheduled: %s", pProposal.String()))

	return nil
}
