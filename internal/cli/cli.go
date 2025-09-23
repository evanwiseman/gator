package cli

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/evanwiseman/gator/internal/config"
	"github.com/evanwiseman/gator/internal/database"
	"github.com/google/uuid"
)

type State struct {
	DB  *database.Queries
	Cfg *config.Config
}

type Command struct {
	Name string
	Args []string
}

func HandlerLogin(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) == 0 {
		return fmt.Errorf("error no username provided")
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("error expects only 1 username")
	}

	// Check user is in database
	context := context.Background()
	userName := cmd.Args[0]
	_, err := s.DB.GetUser(context, sql.NullString{String: userName, Valid: true})
	if err != nil {
		return fmt.Errorf("error user not in database: %v", err)
	}
	err = s.Cfg.SetUserName(userName)
	if err != nil {
		return fmt.Errorf("error couldn't set username: %v", err)
	}
	fmt.Printf("User successfully set to %v\n", userName)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) == 0 {
		return fmt.Errorf("error no username provided")
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("error expects only 1 username")
	}

	context := context.Background()
	userName := cmd.Args[0]
	user, err := s.DB.CreateUser(context, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      sql.NullString{String: userName, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error user '%v' is already registered: %v", userName, err)
	}

	err = s.Cfg.SetUserName(userName)
	if err != nil {
		return fmt.Errorf("error unable to set user name '%v': %v", userName, err)
	}
	fmt.Printf("User '%v' successfully registered\n", userName)
	fmt.Printf("User{\n  %v\n  %v\n  %v  \n  %v\n}", user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
	return nil
}

type Commands struct {
	Registry map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	cmdName := cmd.Name
	cmdHandler, ok := c.Registry[cmdName]
	if !ok {
		return fmt.Errorf("error running command '%v': not in registry", cmdName)
	}
	err := cmdHandler(s, cmd)
	if err != nil {
		return fmt.Errorf("error running command '%v': %v", cmdName, err)
	}
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Registry[name] = f
}
