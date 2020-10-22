package vm_client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/dfinance/dvm-proto/go/compiler_grpc"
	"github.com/dfinance/dvm-proto/go/metadata_grpc"
	"github.com/dfinance/dvm-proto/go/types_grpc"
	"google.golang.org/grpc"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers"
)

const (
	FlagCompilerAddr  = "compiler"
	FlagOutput        = "to-file"
	FlagCompilerUsage = "--compiler " + config.DefaultCompilerAddr
)

// Create connection to vm.
func CreateConnection(addr string) (*grpc.ClientConn, error) {
	return helpers.GetGRpcClientConnection(addr, 0)
}

// Extract arguments from bytecode with compiler.
func ExtractArguments(addr string, bytecode []byte) ([]types_grpc.VMTypeTag, error) {
	conn, err := CreateConnection(addr)
	if err != nil {
		return nil, fmt.Errorf("Can't extract contract metadata because of error during connection to VM: %s\n", err.Error())
	}
	defer conn.Close()

	client := metadata_grpc.NewDVMBytecodeMetadataClient(conn)
	connCtx := context.Background()

	res, err := client.GetMetadata(connCtx, &metadata_grpc.Bytecode{Code: bytecode})
	if err != nil {
		return nil, fmt.Errorf("Can't extract contract metadata because of error during connection to VM: %s\n", err.Error())
	}

	if res.GetScript() == nil {
		return nil, fmt.Errorf("can't extract contract metadata, received not script bytecode")
	}

	return res.GetScript().Arguments, nil
}

// Compile script via grpc compiler.
func Compile(addr string, sourceFiles *compiler_grpc.SourceFiles) ([]CompiledItem, error) {
	conn, err := CreateConnection(addr)
	if err != nil {
		return nil, fmt.Errorf("compilation failed because of error during connection to VM (%s): %w", addr, err)
	}
	defer conn.Close()

	compilerClient := compiler_grpc.NewDvmCompilerClient(conn)
	metadataClient := metadata_grpc.NewDVMBytecodeMetadataClient(conn)
	connCtx := context.Background()

	compResp, err := compilerClient.Compile(connCtx, sourceFiles)
	if err != nil {
		return nil, fmt.Errorf("compilation failed because of error during compilation and connection to VM (%s): %w", addr, err)
	}

	// if contains errors
	if len(compResp.Errors) > 0 {
		return nil, fmt.Errorf("compilation failed because of errors from compiler: %s", strings.Join(compResp.Errors, "\n"))
	}

	resp := make([]CompiledItem, len(compResp.Units))

	for i, unit := range compResp.Units {
		resp[i] = CompiledItem{
			ByteCode: unit.Bytecode,
			Code:     hex.EncodeToString(unit.Bytecode),
		}

		meta, err := metadataClient.GetMetadata(connCtx, &metadata_grpc.Bytecode{Code: unit.Bytecode})
		if err != nil {
			return nil, fmt.Errorf("compilation failed because of error during getting meta information (%s): %w", addr, err)
		}

		if ok := meta.GetScript(); ok != nil {
			resp[i].CodeType = CodeTypeScript
		}

		if moduleMeta := meta.GetModule(); moduleMeta != nil {
			resp[i].CodeType = CodeTypeModule
			resp[i].Types, resp[i].Methods = MatchProtoFields(moduleMeta)
		}
	}

	return resp, nil
}

// MatchProtoFields decorates protobuf structures for using custom fields.
func MatchProtoFields(meta *metadata_grpc.ModuleMeta) (types []ModuleType, methods []ModuleMethod) {
	if meta == nil {
		panic("received metadata is nil")
	}

	methods = make([]ModuleMethod, len(meta.GetFunctions()))
	for i, item := range meta.GetFunctions() {
		methods[i] = ModuleMethod{
			Name:           item.Name,
			IsPublic:       item.IsPublic,
			IsNative:       item.IsNative,
			TypeParameters: item.TypeParameters,
			Arguments:      item.Arguments,
			Returns:        item.Returns,
		}
	}

	types = make([]ModuleType, len(meta.GetTypes()))
	for i, item := range meta.GetTypes() {
		types[i] = ModuleType{
			Name:           item.Name,
			IsResource:     item.IsResource,
			TypeParameters: item.TypeParameters,
			Field:          make([]ModuleTypeField, len(item.Field)),
		}

		for fi, filed := range item.Field {
			types[i].Field[fi].Name = filed.Name
			types[i].Field[fi].Type = filed.Type
		}
	}

	return
}
