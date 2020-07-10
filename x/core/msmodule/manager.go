package msmodule

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// MsManager extends std modules manager with multi signature supported modules.
type MsManager struct {
	*module.Manager
	MsModules map[string]AppMsModule
}

// NewMsManager creates a new multi signature module manager.
func NewMsManager(modules ...interface{}) *MsManager {
	moduleMap := make(map[string]module.AppModule)
	msModulesMap := make(map[string]AppMsModule)

	var modulesStr []string

	for _, instance := range modules {
		if _, ok := instance.(module.AppModule); !ok {
			panic("not an module!")
		}

		realModule := instance.(module.AppModule)

		moduleMap[realModule.Name()] = realModule
		modulesStr = append(modulesStr, realModule.Name())

		if msModule, ok := instance.(AppMsModule); ok {
			msModulesMap[realModule.Name()] = msModule
		}
	}

	return &MsManager{&module.Manager{
		Modules:            moduleMap,
		OrderInitGenesis:   modulesStr,
		OrderExportGenesis: modulesStr,
		OrderBeginBlockers: modulesStr,
		OrderEndBlockers:   modulesStr,
	}, msModulesMap}
}

// RegisterMsRoutes registers multi signature routes.
func (m *MsManager) RegisterMsRoutes(router MsRouter) {
	for _, module := range m.MsModules {
		switch module.(type) {
		case AppMsModule:
			if module.Route() != "" {
				router.AddRoute(module.Route(), module.NewMsHandler())
			}
		}
	}

	router.Seal()
}

// RegisterInvariants registers all module routes and module querier routes.
// {blackList} allows to skip invariants register for specific module names.
func (m *MsManager) RegisterInvariants(ir sdk.InvariantRegistry, blackList... string) {
	for _, module := range m.Modules {
		blackListed := false
		for _, disabledModuleName := range blackList {
			if module.Name() == disabledModuleName {
				blackListed = true
				break
			}
		}

		if !blackListed {
			module.RegisterInvariants(ir)
		}
	}
}
