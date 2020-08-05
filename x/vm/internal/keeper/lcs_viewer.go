package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/dfinance/lcs"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

func StringifyLCSData(request types.ViewerRequest, lscData []byte) (string, error) {
	structType, err := getStructReflectType(request, "")
	if err != nil {
		return "", fmt.Errorf("buidling viewer request: %w", err)
	}
	structPtr := reflect.New(structType)

	if err := lcs.Unmarshal(lscData, structPtr.Interface()); err != nil {
		return "", fmt.Errorf("LCS unmarshal: %w", err)
	}

	output, err := json.MarshalIndent(structPtr.Interface(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("result JSON marshal: %w", err)
	}

	return string(output), nil
}

func getStructReflectType(request types.ViewerRequest, rootFieldName string) (reflect.Type, error) {
	buildErr := func(item types.ViewerItem, msg string) error {
		rootPrefix := ""
		if rootFieldName != "" {
			rootPrefix = rootFieldName + ": "
		}

		return fmt.Errorf("%sfield %q (%s): %s", rootPrefix, item.Name, item.Type, msg)
	}

	var structFields []reflect.StructField
	for _, item := range request {
		if item.Name == "" {
			return nil, buildErr(item, "empty struct field name")
		}

		fieldName := strings.Title(strings.ToLower(item.Name))
		fieldType, err := getFieldReflectType(item)
		if err != nil {
			return nil, buildErr(item, err.Error())
		}

		structFields = append(structFields, reflect.StructField{Name: fieldName, Type: fieldType})
	}
	structType := reflect.StructOf(structFields)

	return structType, nil
}

func getFieldReflectType(item types.ViewerItem) (reflect.Type, error) {
	switch item.Type {
	case types.ViewerTypeU8:
		return reflect.TypeOf(uint8(0)), nil
	case types.ViewerTypeU64:
		return reflect.TypeOf(uint64(0)), nil
	case types.ViewerTypeU128:
		return reflect.TypeOf(&big.Int{}), nil
	case types.ViewerTypeBool:
		return reflect.TypeOf(false), nil
	case types.ViewerTypeAddress:
		return reflect.TypeOf([20]uint8{}), nil
	case types.ViewerTypeStruct:
		if item.InnerItem == nil {
			return nil, fmt.Errorf("inner_item: nil")
		}

		return getStructReflectType(*item.InnerItem, item.Name)
	case types.ViewerTypeVector:
		if item.InnerItem == nil {
			return nil, fmt.Errorf("inner_item: nil")
		}
		if len(*item.InnerItem) != 1 {
			return nil, fmt.Errorf("inner_item: must contain one element")
		}

		sliceType, err := getFieldReflectType((*item.InnerItem)[0])
		if err != nil {
			return nil, fmt.Errorf("inner_item[0]: %w", err)
		}

		return reflect.SliceOf(sliceType), nil
	}

	return nil, fmt.Errorf("unknown field type %q", item.Type)
}
