package vm_client

import (
	"context"
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

// MVFile struct contains code from file in hex.
type MoveFile struct {
	Code string `json:"code"`
}

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

	return res.GetScript().Arguments, nil
}

// Compile script via grpc compiler.
func Compile(addr string, sourceFiles *compiler_grpc.SourceFiles) ([]byte, error) {
	conn, err := CreateConnection(addr)
	if err != nil {
		return nil, fmt.Errorf("compilation failed because of error during connection to VM (%s): %w", addr, err)
	}
	defer conn.Close()

	client := compiler_grpc.NewDvmCompilerClient(conn)
	connCtx := context.Background()

	resp, err := client.Compile(connCtx, sourceFiles)
	if err != nil {
		return nil, fmt.Errorf("compilation failed because of error during compilation and connection to VM (%s): %w", addr, err)
	}

	// if contains errors
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("compilation failed because of errors from compiler: %s", strings.Join(resp.Errors, "\n"))
	}

	return resp.Units[0].Bytecode, nil
}
