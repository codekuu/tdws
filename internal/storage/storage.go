package storage

import (
	"github.com/codekuu/tdws/internal/config"
	"github.com/codekuu/tdws/internal/module"
)

// Returns a list of all modules in the storage with full path
func GetModulePaths(cfg config.Config) []string {
	modules := []string{}
	for _, mod := range cfg.Modules {
		modulePath := module.GetPathFromModule(cfg, mod)
		modules = append(modules, modulePath)
	}
	return modules
}
