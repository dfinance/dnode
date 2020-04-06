// +build integ

package keeper

import (
	"io/ioutil"
	"os"
	"testing"
)

func getGenesis(t *testing.T) []byte {
	fileName := "./genesis_ws.json"

	handle, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("can't read write set: %v", err)
	}
	defer handle.Close()
	bz, err := ioutil.ReadAll(handle)
	if err != nil {
		t.Fatalf("can't read json content of genesis state: %v", err)
	}

	return bz
}
