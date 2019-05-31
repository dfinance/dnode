package helpers

import "encoding/hex"

const (
	EthAddressLength = 20
)

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

func IsEthereumAddress(address string) bool {
	s := address[:2]
	return len(s) == 2*EthAddressLength && isHex(s)
}