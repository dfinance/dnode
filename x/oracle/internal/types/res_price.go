package types

import (
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/glav"
	"github.com/dfinance/lcs"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// ResCurrentPrice is a DVM resource, containing current asset price.
type ResPrice struct {
	Value *big.Int
}

// GetAssetCodePath returns vm_grpc.VMAccessPath for storing price DVM resource.
func GetAssetCodePath(assetCode dnTypes.AssetCode) (*vm_grpc.VMAccessPath, error) {
	assets := strings.Split(assetCode.String(), string(dnTypes.AssetCodeDelimiter))
	if len(assets) != 2 {
		return nil, fmt.Errorf("converting assetCode %q to VMAccessPath: invalid AssetCode", assetCode.String())
	}

	return &vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    glav.OracleAccessVector(assets[0], assets[1]),
	}, nil
}

// NewResPriceStorageValuesPanic returns VM storage key/value for current oracle price DVM resource, panics on error.
func NewResPriceStorageValuesPanic(assetCode dnTypes.AssetCode, price sdk.Int) (*vm_grpc.VMAccessPath, []byte) {
	key, err := GetAssetCodePath(assetCode)
	if err != nil {
		panic(err)
	}

	res := ResPrice{Value: price.BigInt()}
	value, err := lcs.Marshal(res)
	if err != nil {
		panic(fmt.Errorf("oracle ResPrice value for %q (%s) lcs.Marshal: %w", assetCode.String(), price.String(), err))
	}

	return key, value
}
