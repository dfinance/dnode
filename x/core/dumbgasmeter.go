package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DumbGasMeter struct {
}

func NewDumbGasMeter() sdk.GasMeter {
	return &DumbGasMeter{}
}

func (g DumbGasMeter) GasConsumed() sdk.Gas {
	return 0
}

func (g DumbGasMeter) Limit() sdk.Gas {
	return 0
}

func (g DumbGasMeter) GasConsumedToLimit() sdk.Gas {
	return 0
}

func (g *DumbGasMeter) ConsumeGas(_ sdk.Gas, _ string) {
}

func (g DumbGasMeter) IsPastLimit() bool {
	return false
}

func (g DumbGasMeter) IsOutOfGas() bool {
	return false
}
