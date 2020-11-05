package types

import (
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/dfinance/dnode/x/migration/internal/migrations/v1_0"
)

// MigrationHandler converts an appState (genesis map) from the previous version to the targeted one.
type MigrationHandler func(initialAppState genutil.AppMap) (migratedAppState genutil.AppMap, retErr error)

// MigrationMap defines a mapping from a migration target version to a MigrationHandler.
type TargetMigrationMap map[string]MigrationHandler

// MigrationMap is a registered migrations map.
var MigrationMap = TargetMigrationMap{
	"v1.0": v1_0.Migrate,
}
