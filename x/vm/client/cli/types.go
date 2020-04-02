package cli

import vmClient "github.com/dfinance/dnode/x/vm/client"

const (
	FlagOutput        = "to-file"
	FlagCompilerAddr  = "compiler"
	FlagCompilerUsage = "--compiler " + vmClient.DefaultCompilerAddr
)
