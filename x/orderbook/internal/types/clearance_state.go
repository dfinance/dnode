package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ClearanceState object stores the PQCurve crossing point details.
type ClearanceState struct {
	// Crossing point price
	Price sdk.Uint
	// Relation coefficient between crossing point supply and demand (supply / demand)
	ProRata sdk.Dec
	// Inverted ProRata coefficient (1 / ProRata)
	ProRataInvert sdk.Dec
	// Crossing point demand volume adjusted by ProRata (demand * ProRata)
	MaxBidVolume sdk.Dec
	// Crossing point supply volume adjusted by ProRata (supply * ProRataInvert)
	MaxAskVolume sdk.Dec
}

// Strings returns multi-line text object representation.
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
