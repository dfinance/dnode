package cli

import vmClient "github.com/dfinance/dnode/x/vm/client"

const (
	FlagOutput        = "to-file"
	FlagCompilerUsage = "--compiler " + vmClient.DefaultCompilerAddr
)
