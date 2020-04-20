package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Plan struct {
	Height int64 `json:"height"`
}

func (p Plan) String() string {
	return fmt.Sprintf(`Plan:
  blockHeight %d
`, p.Height)
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

func NewPlan(blockHeight int64) Plan {
	return Plan{Height: blockHeight}
}
