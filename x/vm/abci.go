package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

func BeginBlocker(ctx sdk.Context, keeper Keeper, _ abci.RequestBeginBlock) {
	keeper.IterateProposalsQueue(ctx, func(id uint64, p types.ExecutableProposal) bool {
		if !p.Data.GetPlan().ShouldExecute(ctx) {
			return false
		}

		var err error

		switch p.Type {
		case types.ProposalTypeModuleUpdate:
			err = handleModuleUpdateProposalExecution(ctx, keeper, p.Data)
		case types.ProposalTypeTest:
			err = handleTestProposalExecution(ctx, keeper, p.Data)
		default:
			panic(fmt.Errorf("unsupported type: %s", p.String()))
		}

		if err != nil {
			panic(fmt.Errorf("execution failed: %s: %v", p.String(), err))
		}

		keeper.RemoveProposalFromQueue(ctx, id)

		return false
	})
}

func handleModuleUpdateProposalExecution(ctx sdk.Context, keeper Keeper, p types.PlannedProposal) error {
	logger := keeper.Logger(ctx)

	proposal, ok := p.(ModuleUpdateProposal)
	if !ok {
		logger.Error("abci ModuleUpdateProposal: type assert failed")
	}

	logger.Info(fmt.Sprintf("abci ModuleUpdateProposal: executing: %s", proposal.String()))

	return nil
}

func handleTestProposalExecution(ctx sdk.Context, keeper Keeper, p types.PlannedProposal) error {
	logger := keeper.Logger(ctx)

	proposal, ok := p.(TestProposal)
	if !ok {
		logger.Error("abci TestProposal: type assert failed")
	}

	logger.Info(fmt.Sprintf("abci TestProposal: executing: %s", proposal.String()))

	return nil
}