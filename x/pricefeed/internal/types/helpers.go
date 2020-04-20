package types

import (
	"fmt"
	"unicode"
)

func assetCodeFilter(code string) error {
	return stringFilter(code, []strFilterOpt{stringIsEmpty}, []runeFilterOpt{runeIsASCII, runeLetterIsLowerCase})
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

func stringIsEmpty(str string) error {
	if len(str) == 0 {
		return fmt.Errorf("empty")
	}

	return nil
}

func runeIsASCII(rValue rune) error {
	if rValue > unicode.MaxASCII {
		return fmt.Errorf("non ASCII symbol")
	}

	return nil
}

func runeLetterIsLowerCase(rValue rune) error {
	if unicode.IsLetter(rValue) && !unicode.IsLower(rValue) {
		return fmt.Errorf("non lower case symbol")
	}

	return nil
}
