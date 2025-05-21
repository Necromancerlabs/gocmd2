# gocmd2

A Go-based interactive shell framework inspired by Python's [cmd2](https://github.com/python-cmd2/cmd2) library. Gocmd2 provides a modular, extensible framework for building interactive command-line applications with rich features.

## Features

- **Modular Architecture**: Easily create and register command modules to organize functionality
- **Built on Cobra**: Uses [Spf13 Cobra](https://github.com/spf13/cobra) for command parsing and execution
- **Command Auto-Completion**: Tab completion for commands via [Readline](https://github.com/chzyer/readline)
- **Command History**: Persistent command history between sessions
- **Module System**: Enable/disable command modules at runtime
- **Shared State**: Built-in mechanism for sharing state between modules
- **Extensible**: Simple API for adding custom commands and modules
- **Exit Handlers**: Register functions to be called on shell exit

## Installation

```bash
go get github.com/Necromancerlabs/gocmd2
```

## Quick Start

```go
package main

import (
	"fmt"
	"os"

	"github.com/Necromancerlabs/gocmd2/pkg/shell"
	"github.com/spf13/cobra"
)

func main() {
	// Create a new shell
	sh, err := shell.NewShell(
		"myshell",                         // Shell name
		"Welcome to My Interactive Shell!" // Banner message
	)
	if err != nil {
		fmt.Printf("Error initializing shell: %v\n", err)
		os.Exit(1)
	}
	defer sh.Close()

	// Run the shell
	sh.Run()
}
```

This creates a basic shell with core commands like `help`, `exit`, and module management commands.

## Creating Custom Modules

Extend functionality by implementing the `module.CommandModule` interface:

```go
package main

import (
	"fmt"
	"time"

	"github.com/Necromancerlabs/gocmd2/pkg/shellapi"
	"github.com/spf13/cobra"
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
	m.shell.SetState("start_time", time.Now())
}

func (m *TimerModule) GetCommands() []*cobra.Command {
	commands := []*cobra.Command{}

	// Add a "time" command
	timeCmd := &cobra.Command{
		Use:   "time",
		Short: "Show elapsed time since shell started",
		Run: func(cmd *cobra.Command, args []string) {
			startTimeValue, ok := m.shell.GetState("start_time")
			if !ok {
				fmt.Println("Start time not found in state")
				return
			}

			startTime := startTimeValue.(time.Time)
			elapsed := time.Since(startTime)
			fmt.Printf("Shell has been running for %s\n", elapsed.Round(time.Second))
		},
	}
	commands = append(commands, timeCmd)

	return commands
}
```

Register your module in your shell application:

```go
// Register our custom timer module
timerModule := NewTimerModule()
sh.RegisterModule(timerModule)
```

## Running the Examples

The repository includes examples that demonstrate gocmd2's features and usage patterns:

### Timer Example

The `examples/simple` directory contains a working example of a timer module that demonstrates:
- Creating a custom module
- Using shared state between commands
- Dynamically changing the shell prompt
- Setting up exit handlers

To run the example:

```bash
# Clone the repository
git clone https://github.com/Necromancerlabs/gocmd2.git
cd gocmd2

# Run the simple example
go run examples/simple/main.go
```

Once the shell starts, you can try these commands:
- `help` - List available commands
- `time` - Show elapsed time since the shell started
- `reset` - Reset the timer
- `exit` - Exit the shell (with cleanup)

This example shows how to create interactive shells with custom commands and shared state management.

## Core Features

### Module Management

Users can enable/disable modules at runtime:

```
> modules            # List all modules
> enable mymodule    # Enable a specific module
> disable mymodule   # Disable a specific module
```

### Shell API

The Shell API provides methods for modules to interact with the shell:

- **State Management**: `SetState()`, `GetState()`
- **UI Methods**: `SetPrompt()`, `GetPrompt()`, `PrintAlert()`
- **Module Management**: `EnableModule()`, `DisableModule()`, `IsModuleEnabled()`

### Exit Handling

Register cleanup functions to run when the shell exits:

```go
sh.OnExit(func() {
    fmt.Println("Cleaning up resources...")
    // Cleanup code here
})
```

## License

Refer to the LICENSE file for details.
