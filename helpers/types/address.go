package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Parse string address representation (both libra hex and bech32).
func GetAccAddressFromHexOrBech32(strAddr string) (address sdk.AccAddress, err error) {
	address, err = hex.DecodeString(strAddr)
	if err != nil {
		address, err = sdk.AccAddressFromBech32(strAddr)
		if err != nil {
			err = fmt.Errorf("can't parse address %q (should be libra hex or bech32): %v", strAddr, err)
		}
	}

	return
}
