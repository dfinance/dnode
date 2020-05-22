package clitester

import (
	"fmt"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func makeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc) // register all module codecs.
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

func trimCliOutput(output []byte) []byte {
	for i := 0; i < len(output); i++ {
		if output[i] == '{' {
			output = output[i:]
			break
		}
	}

	return output
}

func WaitForFileExists(filePath string, timeoutDur time.Duration) error {
	timeoutCh := time.After(timeoutDur)

	for {
		select {
		case <-timeoutCh:
			return fmt.Errorf("file %q did not appear after %v", filePath, timeoutDur)
		default:
			if _, err := os.Stat(filePath); err == nil {
				return nil
			}
		}
	}
}
