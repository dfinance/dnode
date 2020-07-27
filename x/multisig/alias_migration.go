package multisig

import (
	v06 "github.com/dfinance/dnode/x/multisig/internal/legacy/v0_6"
	v07 "github.com/dfinance/dnode/x/multisig/internal/legacy/v0_7"
)

type (
	GenesisStateV06 = v06.GenesisState
	GenesisStateV07 = v07.GenesisState
)

var (
	MigrateV06ToV07 = v07.Migrate
)
