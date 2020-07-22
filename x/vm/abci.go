package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker handles gov proposal scheduler: iterating over plannedProposals and checking if it is time to execute.
func BeginBlocker(ctx sdk.Context, k Keeper, _ abci.RequestBeginBlock) {
	logger := k.GetLogger(ctx)

	// setup current (actual) DS context
	k.SetDSContext(ctx)

	// proposals processing
	k.IterateProposalsQueue(ctx, func(id uint64, pProposal PlannedProposal) {
		if !pProposal.GetPlan().ShouldExecute(ctx) {
			return
		}

		var err error

		switch proposal := pProposal.(type) {
		case StdlibUpdateProposal:
			err = handleStdlibUpdateProposalExecution(ctx, k, proposal)
		default:
			panic(fmt.Errorf("unsupported type: %T", pProposal))
		}

		if err != nil {
			logger.Error(fmt.Sprintf("%s\nexecution status: failed: %v", pProposal.String(), err))
		} else {
			k.SetDSContext(ctx)
			logger.Info(fmt.Sprintf("%s\nexecution status: done", pProposal.String()))
		}

		k.RemoveProposalFromQueue(ctx, id)
	})
}

// handleStdlibUpdateProposalExecution requests DVM to update stdlib.
func handleStdlibUpdateProposalExecution(ctx sdk.Context, k Keeper, proposal StdlibUpdateProposal) error {
	msg, _ := getStdlibUpdateMsg(proposal)
	if err := k.DeployContract(ctx, msg); err != nil {
		return err
	}

	return nil
}
