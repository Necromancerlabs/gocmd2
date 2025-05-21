package shell

import (
	"fmt"
	"strings"
	"sync"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/Necromancerlabs/gocmd2/pkg/module"
	"github.com/Necromancerlabs/gocmd2/pkg/module/core"
	"github.com/Necromancerlabs/gocmd2/pkg/shellapi"
)

// Shell represents our interactive shell application
type Shell struct {
	rootCmd        *cobra.Command
	rl             *readline.Instance
	currentPrompt  string
	commandModules []module.CommandModule
	banner         string

	// Track which modules are enabled
	enabledModules map[string]bool
	moduleCommands map[string][]*cobra.Command

	// Shared state accessible to all modules
	State      map[string]interface{}
	stateMutex sync.RWMutex
}

// Ensure Shell implements ShellAPI
var _ shellapi.ShellAPI = (*Shell)(nil)

// NewShell creates a new shell instance with core commands pre-registered
func NewShell(rootCmdName, banner string) (*Shell, error) {
	// Use defaults if not provided
	if rootCmdName == "" {
		rootCmdName = "shell"
	}

	shell := &Shell{
		currentPrompt:  "> ",
		banner:         banner,
		State:          make(map[string]interface{}),
		enabledModules: make(map[string]bool),
		moduleCommands: make(map[string][]*cobra.Command),
	}

	// Initialize the root command
	shell.rootCmd = &cobra.Command{
		Use:                   rootCmdName,
		Short:                 "Interactive shell",
		DisableSuggestions:    true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
	}
	shell.rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Initialize readline
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          shell.currentPrompt,
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, err
	}
	shell.rl = rl

	// Register the core module by default
	coreModule := core.New()
	shell.RegisterModule(coreModule)

	return shell, nil
}

// RegisterModule adds a new command module to the shell
func (s *Shell) RegisterModule(module module.CommandModule) {
	moduleName := module.Name()

	// Add this module to our list
	s.commandModules = append(s.commandModules, module)

	// Store the commands for this module
	commands := module.GetCommands()
	s.moduleCommands[moduleName] = commands

	// Enable this module by default
	s.enabledModules[moduleName] = true

	// Add the module's commands to the root command
	for _, cmd := range commands {
		s.rootCmd.AddCommand(cmd)
	}

	// Initialize the module with a reference to the shell
	module.Initialize(s)

	// Update command completion
	s.updateCompleter()
}

// SetPrompt changes the shell prompt
func (s *Shell) SetPrompt(prompt string) {
	s.currentPrompt = prompt + " "
	s.rl.SetPrompt(s.currentPrompt)
}

// GetPrompt returns the current prompt string
func (s *Shell) GetPrompt() string {
	return strings.TrimSpace(s.currentPrompt)
}

// GetReadline returns the readline instance
func (s *Shell) GetReadline() *readline.Instance {
	return s.rl
}

// SetState sets a value in the shared state
func (s *Shell) SetState(key string, value interface{}) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	s.State[key] = value
}

// GetState gets a value from the shared state
func (s *Shell) GetState(key string) (interface{}, bool) {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	val, ok := s.State[key]
	return val, ok
}

// updateCompleter rebuilds the auto-completion based on available commands
func (s *Shell) updateCompleter() {
	completer := readline.NewPrefixCompleter()
	for _, cmd := range s.rootCmd.Commands() {
		completer.Children = append(completer.Children, readline.PcItem(cmd.Name()))
	}
	s.rl.Config.AutoComplete = completer
}

func (s *Shell) PrintAlert(message string) {
	s.rl.Write([]byte(message + "\n"))
	s.rl.Refresh()
}

// Run starts the shell's main loop
func (s *Shell) Run() {
	if s.banner != "" {
		fmt.Println(s.banner)
	} else {
		fmt.Println("Interactive shell started. Type 'help' for available commands.")
	}

	// Main REPL loop
	for {
		line, err := s.rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the line and execute the command using Cobra
		args := strings.Split(line, " ")
		s.rootCmd.SetArgs(args)

		err = s.rootCmd.Execute()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		// Reset rootCmd for next command
		s.rootCmd.SetArgs(nil)
	}
}

// ExecuteCommand runs a command programmatically
func (s *Shell) ExecuteCommand(command string) error {
	args := strings.Split(command, " ")
	s.rootCmd.SetArgs(args)
	err := s.rootCmd.Execute()
	s.rootCmd.SetArgs(nil)
	return err
}

// OnExit registers handlers to be called when the shell exits
func (s *Shell) OnExit(fn func()) {
	// Hook into the exit command
	for _, cmd := range s.rootCmd.Commands() {
		if cmd.Name() == "exit" {
			originalRun := cmd.Run
			cmd.Run = func(c *cobra.Command, args []string) {
				fn()                 // Run the exit handler
				originalRun(c, args) // Call the original exit function
			}
			break
		}
	}
}

// SetHistoryFile changes the history file location
func (s *Shell) SetHistoryFile(path string) error {
	s.rl.SetHistoryPath(path)
	return nil
}

// Close cleans up the shell resources
func (s *Shell) Close() {
	s.rl.Close()
}

// EnableModule enables a module by name, adding its commands to the shell
func (s *Shell) EnableModule(moduleName string) error {
	// Check if module exists
	found := false
	for _, module := range s.commandModules {
		if module.Name() == moduleName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("module not found: %s", moduleName)
	}

	// If already enabled, do nothing
	if s.enabledModules[moduleName] {
		return nil
	}

	// Enable the module
	s.enabledModules[moduleName] = true

	// Add the module's commands to the root command
	for _, cmd := range s.moduleCommands[moduleName] {
		s.rootCmd.AddCommand(cmd)
	}

	// Update completer
	s.updateCompleter()
	return nil
}

// DisableModule disables a module by name, removing its commands from the shell
func (s *Shell) DisableModule(moduleName string) error {
	// Don't allow disabling the core module
	if moduleName == "core" {
		return fmt.Errorf("cannot disable core module")
	}

	// Check if module exists
	found := false
	for _, module := range s.commandModules {
		if module.Name() == moduleName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("module not found: %s", moduleName)
	}

	// If already disabled, do nothing
	if !s.enabledModules[moduleName] {
		return nil
	}

	// Disable the module
	s.enabledModules[moduleName] = false

	// Remove the module's commands from the root command
	for _, cmd := range s.moduleCommands[moduleName] {
		s.rootCmd.RemoveCommand(cmd)
	}

	// Update completer
	s.updateCompleter()
	return nil
}

// IsModuleEnabled returns whether a module is enabled
func (s *Shell) IsModuleEnabled(moduleName string) bool {
	return s.enabledModules[moduleName]
}

// GetModules returns a list of all module names
func (s *Shell) GetModules() []string {
	modules := make([]string, 0, len(s.commandModules))
	for _, module := range s.commandModules {
		modules = append(modules, module.Name())
	}
	return modules
}

// GetEnabledModules returns a list of enabled module names
func (s *Shell) GetEnabledModules() []string {
	enabled := []string{}
	for name, isEnabled := range s.enabledModules {
		if isEnabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// GetRootCmd returns the shell's root command
func (s *Shell) GetRootCmd() *cobra.Command {
	return s.rootCmd
}

// GetModuleCommands returns a map of module names to their commands
func (s *Shell) GetModuleCommands() map[string][]*cobra.Command {
	return s.moduleCommands
}
