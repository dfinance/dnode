package utils

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewIntFromString(str string, precision uint) (i sdk.Int, err error) {
	if len(str) == 0 {
		return i, errors.New("decimal string is empty")
	}

	// first extract any negative symbol
	neg := false
	if str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return i, errors.New("decimal string is empty")
	}

	strs := strings.Split(str, ".")
	lenDecs := 0
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		lenDecs = len(strs[1])
		if lenDecs == 0 || len(combinedStr) == 0 {
			return i, errors.New("bad decimal length")
		}
		combinedStr = combinedStr + strs[1]

	} else if len(strs) > 2 {
		return i, errors.New("too many periods to be a decimal string")
	}

	if lenDecs > int(precision) {
		return i, errors.New(
			fmt.Sprintf("too much precision, maximum %v, len decimal %v", precision, lenDecs))
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := int(precision) - lenDecs
	zeros := fmt.Sprintf(`%0`+strconv.Itoa(zerosToAdd)+`s`, "")
	combinedStr = combinedStr + zeros

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return i, errors.New(fmt.Sprintf("bad string to integer conversion, combinedStr: %v", combinedStr))
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}
	return sdk.NewIntFromBigInt(combined), nil
}
