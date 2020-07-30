package types

import (
	"fmt"
	"strings"
	"unicode"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AssetCodeDelimiter = '_'
)

func DenomFilter(denom string) error {
	return stringFilter(
		denom,
		[]strFilterOpt{stringNotEmpty, validCoinDenom},
		[]runeFilterOpt{runeIsASCII, runeIsLowerCasedLetter},
	)
}

func AssetCodeFilter(code string) error {
	return stringFilter(
		code,
		[]strFilterOpt{stringNotEmpty, newDelimiterStrFilterOpt(string(AssetCodeDelimiter))},
		[]runeFilterOpt{runeIsASCII, newIsLowerCasedLetterOrDelimiter(AssetCodeDelimiter)},
	)
}

type strFilterOpt func(str string) error

type runeFilterOpt func(rValue rune) error

func stringFilter(str string, strOpts []strFilterOpt, runeOpts []runeFilterOpt) error {
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

func stringNotEmpty(str string) error {
	if len(str) == 0 {
		return fmt.Errorf("empty")
	}

	return nil
}

func validCoinDenom(str string) error {
	if err := sdk.ValidateDenom(str); err != nil {
		return fmt.Errorf("invalid denom: %w", err)
	}

	return nil
}

func newDelimiterStrFilterOpt(delimiter string) strFilterOpt {
	return func(str string) error {
		if strings.HasPrefix(str, delimiter) {
			return fmt.Errorf("delimiter: is a prefix")
		}
		if strings.HasSuffix(str, delimiter) {
			return fmt.Errorf("delimiter: is a suffix")
		}

		n := strings.Count(str, delimiter)
		if n == 0 {
			return fmt.Errorf("delimiter: not found")
		}
		if n > 1 {
			return fmt.Errorf("delimiter: multiple")
		}

		return nil
	}
}

func runeIsASCII(rValue rune) error {
	if rValue > unicode.MaxASCII {
		return fmt.Errorf("non ASCII symbol")
	}

	return nil
}

func runeIsLowerCasedLetter(rValue rune) error {
	if !unicode.IsLetter(rValue) || !unicode.IsLower(rValue) {
		return fmt.Errorf("non lower cased letter symbol")
	}

	return nil
}

func runeLetterIsLowerCase(rValue rune) error {
	if unicode.IsLetter(rValue) && !unicode.IsLower(rValue) {
		return fmt.Errorf("letter symbol is not lower cased")
	}

	return nil
}

func newIsLowerCasedLetterOrDelimiter(delimiter rune) runeFilterOpt {
	return func(rValue rune) error {
		if rValue == delimiter {
			return nil
		}

		return runeIsLowerCasedLetter(rValue)
	}
}
