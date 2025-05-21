// Package shellapi defines interfaces for interactions between the shell and modules
package shellapi

import "github.com/spf13/cobra"

// ShellAPI defines the interface that modules can use to interact with the shell
type ShellAPI interface {
	// Command and module management
	EnableModule(moduleName string) error
	DisableModule(moduleName string) error
	IsModuleEnabled(moduleName string) bool
	GetModules() []string
	GetEnabledModules() []string
	GetRootCmd() *cobra.Command
	GetModuleCommands() map[string][]*cobra.Command

	// Shell state
	SetState(key string, value interface{})
	GetState(key string) (interface{}, bool)

	// UI methods
	SetPrompt(prompt string)
	GetPrompt() string
	PrintAlert(message string)
}
