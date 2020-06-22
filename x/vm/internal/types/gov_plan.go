package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// PlannedProposal is interface for all VM module proposals.
type PlannedProposal interface {
	fmt.Stringer
	GetPlan() Plan
}

// Plan is an object used for proposal scheduling.
type Plan struct {
	Height int64 `json:"height"`
}

func (p Plan) String() string {
	b := strings.Builder{}
	b.WriteString("Plan:\n")
	b.WriteString(fmt.Sprintf("  BlockHeight: %d\n", p.Height))

	return b.String()
}

func (p Plan) ValidateBasic() error {
	if p.Height <= 0 {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, "height can't be <= 0")
	}

	return nil
}

func (p Plan) ShouldExecute(ctx sdk.Context) bool {
	if p.Height > 0 {
		return ctx.BlockHeight() >= p.Height
	}

	return false
}

// NewPlan creates a Plan object.
func NewPlan(blockHeight int64) Plan {
	return Plan{Height: blockHeight}
}
