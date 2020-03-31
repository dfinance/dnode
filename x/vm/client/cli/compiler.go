package cli

import (
	connContext "context"
	"fmt"
	"strings"

	"google.golang.org/grpc"

	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

const (
	FlagOutput          = "to-file"
	FlagCompilerAddr    = "compiler"
	FlagCompilerDefault = "127.0.0.1:50053"
	FlagCompilerUsage   = "--compiler 127.0.0.1:50053"
)

// MVFile struct contains code from file in hex.
type MVFile struct {
	Code string `json:"code"`
}

// Create connection to vm.
func CreateConnection(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, grpc.WithInsecure())
}

// Extract arguments from bytecode with compiler.
func ExtractArguments(addr string, bytecode []byte) ([]vm_grpc.VMTypeTag, error) {
	conn, err := CreateConnection(addr)
	if err != nil {
		return nil, fmt.Errorf("Can't extract contract metadata because of error during connection to VM: %s\n", err.Error())
	}
	defer conn.Close()

	client := vm_grpc.NewVMScriptMetadataClient(conn)
	connCtx := connContext.Background()

	res, err := client.GetSignature(connCtx, &vm_grpc.VMScript{
		Code: bytecode,
	})

	if err != nil {
		return nil, fmt.Errorf("Can't extract contract metadata because of error during connection to VM: %s\n", err.Error())
	}

	return res.Arguments, nil
}

// Compile script via grpc compiler.
func Compile(addr string, sourceFile *vm_grpc.MvIrSourceFile) ([]byte, error) {
	conn, err := CreateConnection(addr)
	if err != nil {
		return nil, fmt.Errorf("compilation failed because of error during connection to VM: %v", err)
	}
	defer conn.Close()

	client := vm_grpc.NewVMCompilerClient(conn)
	connCtx := connContext.Background()

	resp, err := client.Compile(connCtx, sourceFile)
	if err != nil {
		return nil, fmt.Errorf("compilation failed because of error during compilation and connection to VM: %v", err)
	}

	// if contains errors
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("compilation failed because of errors from compiler: %s", strings.Join(resp.Errors, "\n"))
	}

	return resp.Bytecode, nil
}
