package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ClearanceState struct {
	Price         sdk.Uint
	ProRata       sdk.Dec
	ProRataInvert sdk.Dec
	MaxBidVolume  sdk.Dec
	MaxAskVolume  sdk.Dec
}

func (s ClearanceState) String() string {
	b := strings.Builder{}
	b.WriteString("ClearanceState:\n")
	b.WriteString(fmt.Sprintf("  Price:         %s\n", s.Price.String()))
	b.WriteString(fmt.Sprintf("  ProRata:       %s\n", s.ProRata.String()))
	b.WriteString(fmt.Sprintf("  ProRataInvert: %s\n", s.ProRataInvert.String()))
	b.WriteString(fmt.Sprintf("  MaxBidVolume:  %s\n", s.MaxBidVolume.String()))
	b.WriteString(fmt.Sprintf("  MaxAskVolume:  %s\n", s.MaxAskVolume.String()))

	return b.String()
}
