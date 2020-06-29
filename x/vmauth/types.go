package vmauth

import (
	"encoding/hex"
)

var (
	// denomPaths defines matching between denom and path.
	denomPaths map[string][]byte
)

// initialize denomPaths map.
func init() {
	denomPaths = make(map[string][]byte)

	err := AddDenomPath("dfi", "01608540feb9c6bd277405cfdc0e9140c1431f236f7d97865575e830af3dd67e7e")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("eth", "0138f4f2895881c804de0e57ced1d44f02e976f9c6561c889f7b7eef8e660d2c9a")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("usdt", "01a04b6467f35792e0fda5638a509cc807b3b289a4e0ea10794c7db5dc1a63d481")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("btc", "019a2b233aea4cab2e5b6701280f8302be41ea5731af93858fd96e038499eda072")
	if err != nil {
		panic(err)
	}
}

// AddDenomPath add denom/path matching to the denomPaths map.
func AddDenomPath(denom string, path string) error {
	var err error
	denomPaths[denom], err = hex.DecodeString(path)

	return err
}

// RemoveDenomPath removes denom/path matching from the denomPaths map.
func RemoveDenomPath(denom string) {
	delete(denomPaths, denom)
}
