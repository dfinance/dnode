package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker handles gov proposal scheduler: iterating over plannedProposals and checking if it is time to execute.
func BeginBlocker(ctx sdk.Context, keeper Keeper, _ abci.RequestBeginBlock) {
	logger := keeper.Logger(ctx)

	keeper.IterateProposalsQueue(ctx, func(id uint64, pProposal PlannedProposal) {
		if !pProposal.GetPlan().ShouldExecute(ctx) {
			return
		}

		var err error

		switch proposal := pProposal.(type) {
		case StdlibUpdateProposal:
			err = handleStdlibUpdateProposalExecution(ctx, keeper, proposal)
		default:
			panic(fmt.Errorf("unsupported type: %T", pProposal))
		}

		if err != nil {
			logger.Error(fmt.Sprintf("%s\nexecution status: failed: %v", pProposal.String(), err))
		} else {
			keeper.SetDSContext(ctx)
			logger.Info(fmt.Sprintf("%s\nexecution status: done", pProposal.String()))
		}

		keeper.RemoveProposalFromQueue(ctx, id)
	})
}

// handleStdlibUpdateProposalExecution requests DVM to update stdlib.
func handleStdlibUpdateProposalExecution(ctx sdk.Context, keeper Keeper, proposal StdlibUpdateProposal) error {
	msg, _ := getStdlibUpdateMsg(proposal)
	if err := keeper.DeployContract(ctx, msg); err != nil {
		return err
	}

	return nil
}
