package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func BeginBlocker(ctx sdk.Context, keeper Keeper, _ abci.RequestBeginBlock) {
	keeper.IterateProposalsQueue(ctx, func(id uint64, pProposal PlannedProposal) bool {
		if !pProposal.Plan.ShouldExecute(ctx) {
			return false
		}

		var err error

		switch proposal := pProposal.Proposal.(type) {
		case ModuleUpdateProposal:
			err = handleModuleUpdateProposalExecution(ctx, keeper, proposal, pProposal.Data)
		case TestProposal:
			err = handleTestProposalExecution(ctx, keeper, proposal, pProposal.Data)
		default:
			panic(fmt.Errorf("unsupported type: %T", pProposal.Proposal))
		}

		if err != nil {
			panic(fmt.Errorf("execution failed: %s: %v", pProposal.String(), err))
		}

		keeper.RemoveProposalFromQueue(ctx, id)

		return false
	})
}

func handleModuleUpdateProposalExecution(ctx sdk.Context, keeper Keeper, proposal ModuleUpdateProposal, data ProposalData) error {
	logger := keeper.Logger(ctx)

	logger.Info(fmt.Sprintf("abci ModuleUpdateProposal: executing: %s", proposal.String()))

	return nil
}

func handleTestProposalExecution(ctx sdk.Context, keeper Keeper, proposal TestProposal, data ProposalData) error {
	logger := keeper.Logger(ctx)

	logger.Info(fmt.Sprintf("abci TestProposal: executing: %s", proposal.String()))

	return nil
}
