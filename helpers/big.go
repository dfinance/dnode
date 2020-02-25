package helpers

import sdk "github.com/cosmos/cosmos-sdk/types"

// Convert sdk.Int to bytes with little endian by required bit len
func BigToBytes(val sdk.Int, bytesLen int) []byte {
	bytes := val.BigInt().Bytes()

	if len(bytes) < bytesLen {
		diff := bytesLen - len(bytes)
		for i := 0; i < diff; i++ {
			bytes = append([]byte{0}, bytes...)
		}
	}

	for i := 0; i < len(bytes)/2; i++ {
		bytes[i], bytes[len(bytes)-i-1] = bytes[len(bytes)-i-1], bytes[i]
	}

	return bytes
}
