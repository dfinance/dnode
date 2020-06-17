package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// BeginBlocker handles gov proposal scheduler: iterating over plannedProposals and checking if it is time to execute.
func BeginBlocker(ctx sdk.Context, keeper Keeper, _ abci.RequestBeginBlock) {
	keeper.IterateProposalsQueue(ctx, func(id uint64, pProposal PlannedProposal) bool {
		if !pProposal.GetPlan().ShouldExecute(ctx) {
			return false
		}

		var err error

		switch proposal := pProposal.(type) {
		case StdlibUpdateProposal:
			err = handleStdlibUpdateProposalExecution(ctx, keeper, proposal)
		default:
			panic(fmt.Errorf("unsupported type: %T", pProposal))
		}

		if err != nil {
			panic(fmt.Errorf("%s\nexecution failed: %v", pProposal.String(), err))
		}

		keeper.RemoveProposalFromQueue(ctx, id)

		return false
	})
}

// handleStdlibUpdateProposalExecution requests DVM to update stdlib.
func handleStdlibUpdateProposalExecution(ctx sdk.Context, keeper Keeper, proposal StdlibUpdateProposal) error {
	logger := keeper.Logger(ctx)

	msg := types.NewMsgDeployModule(common_vm.ZeroAddress, proposal.Code)
	if err := keeper.DeployContract(ctx, msg); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("proposal executed:\n%s", proposal.String()))

	return nil
}
