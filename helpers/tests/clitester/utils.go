package clitester

import (
	"strings"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
	"k8s.io/kubernetes/pkg/util/slice"
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

// PrintEvents reads webSockets channels and prints sorted attributes.
// If {keyFilters} passed, all attributes are filtered out.
func PrintEvents(t *testing.T, inChs []<-chan coreTypes.ResultEvent, keyFilters ...string) {
	eventWg := sync.WaitGroup{}
	eventCh := make(chan coreTypes.ResultEvent, 100)

	for _, inCh := range inChs {
		eventWg.Add(1)
		go func(ch <-chan coreTypes.ResultEvent) {
			defer eventWg.Done()
			for event := range ch {
				eventCh <- event
			}
		}(inCh)
	}
	go func() {
		eventWg.Wait()
		close(eventCh)
	}()

	for event := range eventCh {
		t.Logf("Got events for query: %s", event.Query)
		keys := make([]string, 0, len(event.Events))
		for key := range event.Events {
			keys = append(keys, key)
		}
		slice.SortStrings(keys)

		for _, key := range keys {
			doPrint := true
			if len(keyFilters) > 0 {
				doPrint = false
				for _, keyFilter := range keyFilters {
					if strings.Contains(key, keyFilter) {
						doPrint = true
					}
				}
			}

			if doPrint {
				t.Logf("  %s: [%s]", key, strings.Join(event.Events[key], ", "))
			}
		}
	}
}
