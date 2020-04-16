package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

// New governance proposal handler for Gov module.
func NewGovHandler(keeper Keeper) gov.Handler {
	return func(ctx sdk.Context, c govTypes.Content) error {
		if c.ProposalRoute() != ModuleName {
			return fmt.Errorf("invalid proposal route %q for module %q", c.ProposalRoute(), ModuleName)
		}

		switch p := c.(type) {
		case TestProposal:
			ctx.Logger().Info(fmt.Sprintf("got VM proposal: %s", p.String()))

			path := vm_grpc.VMAccessPath{
				Address: []byte("my_addr"),
				Path:    []byte("my_path"),
			}

			value := keeper.GetValue(ctx, &path)
			if value == nil {
				ctx.Logger().Info("value not found, setting it")
				keeper.SetValue(ctx, &path, []byte("tst_value"))
			} else {
				ctx.Logger().Info("value found (should not happen)")
			}
		default:
			return fmt.Errorf("unsupported proposal content type %q for module %q", c.ProposalType(), ModuleName)
		}

		return nil
	}
}
