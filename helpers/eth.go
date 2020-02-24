package helpers

import (
	"encoding/hex"
)

const (
	EthAddressLength = 20
)

// Check it's hex
func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// Check if it's ethereum address.
func IsEthereumAddress(address string) bool {
	s := address[2:]
	return len(s) == 2*EthAddressLength && isHex(s)
}
