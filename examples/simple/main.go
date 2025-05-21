package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/Necromancerlabs/gocmd2/pkg/shell"
	"github.com/Necromancerlabs/gocmd2/pkg/shellapi"
)

// TimerModule provides time-related commands
type TimerModule struct {
	shell shellapi.ShellAPI
}

func NewTimerModule() *TimerModule {
	return &TimerModule{}
}

func (m *TimerModule) Name() string {
	return "timer"
}

// Initialize is called when the module is registered
func (m *TimerModule) Initialize(s shellapi.ShellAPI) {
	m.shell = s

	// Store the start time in shared state
	m.shell.SetState("start_time", time.Now())
}

func (m *TimerModule) GetCommands() []*cobra.Command {
	commands := []*cobra.Command{}

	// Time command - shows elapsed time since shell started
	timeCmd := &cobra.Command{
		Use:   "time",
		Short: "Show elapsed time since shell started",
		Run: func(cmd *cobra.Command, args []string) {
			// Get the start time from shared state
			startTimeValue, ok := m.shell.GetState("start_time")
			if !ok {
				fmt.Println("Start time not found in state")
				return
			}

			startTime := startTimeValue.(time.Time)
			elapsed := time.Since(startTime)

			fmt.Printf("Shell has been running for %s\n", elapsed.Round(time.Second))

			// Update the prompt to show running time
			m.shell.SetPrompt(fmt.Sprintf("(%s)", elapsed.Round(time.Second)))
		},
	}
	commands = append(commands, timeCmd)

	// Reset command - resets the timer
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the timer",
		Run: func(cmd *cobra.Command, args []string) {
			m.shell.SetState("start_time", time.Now())
			fmt.Println("Timer reset")
			m.shell.SetPrompt(">")
		},
	}
	commands = append(commands, resetCmd)

	return commands
}

func main() {
	// Create a new shell (core commands are registered automatically)
	sh, err := shell.NewShell("timer-demo", "Welcome to the Timer Demo Shell! Type 'help' for available commands.")
	if err != nil {
		fmt.Printf("Error initializing shell: %v\n", err)
		os.Exit(1)
	}
	defer sh.Close()

	// Register our custom timer module
	timerModule := NewTimerModule()
	sh.RegisterModule(timerModule)

	// Set exit handler
	sh.OnExit(func() {
		fmt.Println("Cleaning up resources...")
		// Example of cleanup code that would be run on exit
	})

	// Run the shell
	sh.Run()
}
