package types

import (
	"fmt"
	"unicode"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ValidateDenom(denom string) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}
	if err := AssetCodeFilter(denom); err != nil {
		return err
	}

	return nil
}

func AssetCodeFilter(code string) error {
	return StringFilter(code, []StrFilterOpt{StringIsEmpty}, []RuneFilterOpt{RuneIsASCII, RuneLetterIsLowerCase})
}

type StrFilterOpt func(str string) error

type RuneFilterOpt func(rValue rune) error

func StringFilter(str string, strOpts []StrFilterOpt, runeOpts []RuneFilterOpt) error {
	for _, opt := range strOpts {
		if err := opt(str); err != nil {
			return err
		}
	}

	if len(runeOpts) == 0 {
		return nil
	}

	for i, r := range str {
		for _, opt := range runeOpts {
			if err := opt(r); err != nil {
				return fmt.Errorf("rune %q at %d index: %w", r, i, err)
			}
		}
	}

	return nil
}

func StringIsEmpty(str string) error {
	if len(str) == 0 {
		return fmt.Errorf("empty")
	}

	return nil
}

func RuneIsASCII(rValue rune) error {
	if rValue > unicode.MaxASCII {
		return fmt.Errorf("non ASCII symbol")
	}

	return nil
}

func RuneLetterIsLowerCase(rValue rune) error {
	if unicode.IsLetter(rValue) && !unicode.IsLower(rValue) {
		return fmt.Errorf("non lower case symbol")
	}

	return nil
}
