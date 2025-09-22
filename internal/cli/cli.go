package cli

import (
	"fmt"

	"github.com/evanwiseman/gator/internal/config"
)

type State struct {
	Config *config.Config
}

type Command struct {
	Name string
	Args []string
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("error 'login' no username provided")
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("error 'login' expects only 1 username")
	}

	username := cmd.Args[0]
	err := s.Config.SetUsername(username)
	if err != nil {
		return fmt.Errorf("error 'login' couldn't set username: %v", err)
	}
	fmt.Printf("Username successfully set to %v\n", username)
	return nil
}

type Commands struct {
	Registry map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	name := cmd.Name
	handler, ok := c.Registry[name]
	if !ok {
		return fmt.Errorf("error running command '%v': not in registry", name)
	}
	err := handler(s, cmd)
	if err != nil {
		return fmt.Errorf("error running command '%v': %v", name, err)
	}
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Registry[name] = f
}
