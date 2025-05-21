package module

import (
	"github.com/spf13/cobra"
	"github.com/Necromancerlabs/gocmd2/pkg/shellapi"
)

// CommandModule interface for adding new command modules
type CommandModule interface {
	// GetCommands returns the cobra commands this module provides
	GetCommands() []*cobra.Command
	// Name returns the name of this module
	Name() string
	// Initialize is called when the module is registered
	Initialize(shell shellapi.ShellAPI)
}
