package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/evanwiseman/gator/internal/cli"
	"github.com/evanwiseman/gator/internal/config"
	"github.com/evanwiseman/gator/internal/database"
	_ "github.com/lib/pq"
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
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	// Store the config state as context
	context := cli.State{
		DB:  dbQueries,
		Cfg: &cfg,
	}

	// Create a command registry and register relevant commands
	commands := cli.Commands{
		Registry: make(map[string]func(*cli.State, cli.Command) error),
	}
	commands.Register("login", cli.HandlerLogin)
	commands.Register("register", cli.HandlerRegister)
	commands.Register("reset", cli.HandlerReset)
	commands.Register("users", cli.HandlerUsers)
	commands.Register("agg", cli.HandlerAgg)
	commands.Register("addfeed", cli.MiddlewareLoggedIn(cli.HandlerAddFeed))
	commands.Register("feeds", cli.HandlerFeeds)
	commands.Register("follow", cli.MiddlewareLoggedIn(cli.HandlerFollow))
	commands.Register("following", cli.MiddlewareLoggedIn(cli.HandlerFollowing))
	commands.Register("unfollow", cli.MiddlewareLoggedIn(cli.HandlerUnfollow))

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
