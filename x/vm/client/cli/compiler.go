package cli

import (
	connContext "context"
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/vm_grpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
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

// Create connection to virtual machine.
func createVMConn() (*grpc.ClientConn, error) {
	return grpc.Dial(viper.GetString(FlagCompilerAddr), grpc.WithInsecure())
}

// Extract arguments from bytecode with compiler.
func extractArgs(bytecode []byte) ([]vm_grpc.VMTypeTag, error) {
	conn, err := createVMConn()
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
func compile(sourceFile *vm_grpc.MvIrSourceFile) ([]byte, bool) {
	conn, err := createVMConn()
	if err != nil {
		fmt.Printf("Compilation failed because of error during connection to VM: %s\n", err.Error())
		return nil, false
	}
	defer conn.Close()

	client := vm_grpc.NewVMCompilerClient(conn)
	connCtx := connContext.Background()

	resp, err := client.Compile(connCtx, sourceFile)
	if err != nil {
		fmt.Printf("Compilation failed because of error during compilation and connection to VM: %s\n", err.Error())
		return nil, false
	}

	// if contains errors
	if len(resp.Errors) > 0 {
		for _, err := range resp.Errors {
			fmt.Printf("Error from compiler: %s\n", err)
		}
		fmt.Println("Compilation failed because of errors from compiler.")
		return nil, false
	}

	return resp.Bytecode, true
}
