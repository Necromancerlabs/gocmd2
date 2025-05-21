package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/Necromancerlabs/gocmd2/pkg/shellapi"
)

// Module provides the essential shell commands
type Module struct {
	shell shellapi.ShellAPI
}

// New creates a new core module
func New() *Module {
	return &Module{}
}

// Name returns the module name
func (m *Module) Name() string {
	return "core"
}

// GetCommands returns all commands provided by this module
func (m *Module) GetCommands() []*cobra.Command {
	commands := []*cobra.Command{}

	// Exit command
	exitCmd := &cobra.Command{
		Use:   "exit",
		Short: "Exit the shell",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Goodbye!")
			os.Exit(0)
		},
	}
	commands = append(commands, exitCmd)

	// Modules command - list all modules
	modulesCmd := &cobra.Command{
		Use:   "modules",
		Short: "List available modules",
		Run: func(cmd *cobra.Command, args []string) {
			modules := m.shell.GetModules()
			fmt.Println("Available modules:")
			for _, name := range modules {
				enabled := m.shell.IsModuleEnabled(name)
				status := "enabled"
				if !enabled {
					status = "disabled"
				}
				fmt.Printf("  %-15s [%s]\n", name, status)
			}
		},
	}
	commands = append(commands, modulesCmd)

	// Enable module command
	enableCmd := &cobra.Command{
		Use:   "enable [module]",
		Short: "Enable a module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]
			err := m.shell.EnableModule(moduleName)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Module '%s' enabled\n", moduleName)
		},
	}
	commands = append(commands, enableCmd)

	// Disable module command
	disableCmd := &cobra.Command{
		Use:   "disable [module]",
		Short: "Disable a module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]
			err := m.shell.DisableModule(moduleName)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Module '%s' disabled\n", moduleName)
		},
	}
	commands = append(commands, disableCmd)

	return commands
}

// InitializeHelp configures the custom help for the shell
func (m *Module) InitializeHelp() {
	// Store the default help function so we can call it later
	defaultHelpFunc := m.shell.GetRootCmd().HelpFunc()

	// Configure the shell's root command help
	m.shell.GetRootCmd().SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// If we're asking for help on the root command with no arguments, show our module-based help
		if cmd == m.shell.GetRootCmd() && len(args) == 0 {
			// Group commands by module
			modules := make(map[string][]*cobra.Command)

			// First, gather all commands by module
			// Use moduleCommands from the shell which tracks which commands belong to which module
			for moduleName, commands := range m.shell.GetModuleCommands() {
				if !m.shell.IsModuleEnabled(moduleName) {
					continue
				}

				cmdList := []*cobra.Command{}
				for _, command := range commands {
					// Look for the command in the root command's commands
					for _, rootCmd := range m.shell.GetRootCmd().Commands() {
						if rootCmd.Name() == command.Name() {
							cmdList = append(cmdList, rootCmd)
							break
						}
					}
				}

				modules[moduleName] = cmdList
			}

			// Special case: Add help command to core module
			for _, rootCmd := range m.shell.GetRootCmd().Commands() {
				if rootCmd.Name() == "help" {
					coreCommands := modules["core"]
					modules["core"] = append(coreCommands, rootCmd)
					break
				}
			}

			// Display commands by module
			fmt.Println("Available commands:")
			enabledModules := m.shell.GetEnabledModules()

			for _, moduleName := range enabledModules {
				cmds, ok := modules[moduleName]
				if !ok || len(cmds) == 0 {
					continue
				}
				fmt.Printf("\n[%s]\n", moduleName)
				for _, cmd := range cmds {
					fmt.Printf("  %-15s %s\n", cmd.Name(), cmd.Short)
				}
			}

			// List any disabled modules at the end
			hasDisabledModules := false
			for _, moduleName := range m.shell.GetModules() {
				if !m.shell.IsModuleEnabled(moduleName) {
					if !hasDisabledModules {
						fmt.Println("\nDisabled modules:")
						hasDisabledModules = true
					}
					fmt.Printf("  %s\n", moduleName)
				}
			}
		} else if len(args) > 0 {
			// If we have arguments, we're looking for help on a specific command
			cmdName := args[0]
			found := false

			// Find the command and print its help
			for _, subCmd := range m.shell.GetRootCmd().Commands() {
				if subCmd.Name() == cmdName || subCmd.NameAndAliases() == cmdName {
					fmt.Printf("Command: %s\n", subCmd.Name())
					fmt.Printf("Usage: %s\n", subCmd.Use)
					if subCmd.Short != "" {
						fmt.Printf("\n%s\n", subCmd.Short)
					}
					if subCmd.Long != "" {
						fmt.Printf("\n%s\n", subCmd.Long)
					}
					if len(subCmd.Aliases) > 0 {
						fmt.Printf("\nAliases: %s\n", strings.Join(subCmd.Aliases, ", "))
					}
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("Unknown command: %s\n", cmdName)
			}
		} else {
			// For any other case, use the default help
			defaultHelpFunc(cmd, args)
		}
	})
}

// Initialize sets up the module
func (m *Module) Initialize(s shellapi.ShellAPI) {
	m.shell = s
	m.InitializeHelp()
}
