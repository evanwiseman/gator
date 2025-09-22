package main

import (
	"fmt"
	"os"

	"github.com/evanwiseman/gator/internal/cli"
	"github.com/evanwiseman/gator/internal/config"
)

func main() {
	// If no arguments are provided exit
	if len(os.Args) < 2 {
		fmt.Println("error no arguments provided")
		os.Exit(1)
	}

	// Read in the config
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// Store the config state as context
	context := cli.State{
		Config: &cfg,
	}

	// Create a command registry and register relevant commands
	commands := cli.Commands{
		Registry: make(map[string]func(*cli.State, cli.Command) error),
	}
	commands.Register("login", cli.HandlerLogin)

	// Create a command from the user provided args and run it with given context
	command := cli.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}
	err = commands.Run(&context, command)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
