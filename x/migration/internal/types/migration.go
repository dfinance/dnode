package types

import (
	"github.com/cosmos/cosmos-sdk/x/genutil"

	v07 "github.com/dfinance/dnode/x/migration/internal/migrations/v0_7"
)

// MigrationHandler converts an appState (genesis map) from the previous version to the targeted one.
type MigrationHandler func(initialAppState genutil.AppMap) (migratedAppState genutil.AppMap, retErr error)

// MigrationMap defines a mapping from a migration target version to a MigrationHandler.
type TargetMigrationMap map[string]MigrationHandler

// MigrationMap is a registered migrations map.
var MigrationMap = TargetMigrationMap{
	"v0.7": v07.Migrate,
}
