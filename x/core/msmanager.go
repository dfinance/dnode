package core

import (
	"github.com/cosmos/cosmos-sdk/types/module"
)

type MsManager struct {
	*module.Manager
	MsModules map[string]AppMsModule
}

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

func (m *MsManager) RegisterMsRoutes(router Router) {
	for _, module := range m.MsModules {
		switch module.(type) {
		case AppMsModule:
			if module.Route() != "" {
				router.AddRoute(module.Route(), module.NewMsHandler())
			}
		}
	}
}
