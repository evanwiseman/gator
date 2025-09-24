package cli

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/evanwiseman/gator/internal/config"
	"github.com/evanwiseman/gator/internal/database"
	"github.com/evanwiseman/gator/internal/rss"
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

	// Set the user in the database
	err = s.Cfg.SetUserName(userName)
	if err != nil {
		return fmt.Errorf("error couldn't set username: %v", err)
	}
	fmt.Printf("user successfully set to %v\n", userName)
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

	// Attempt to create a new user
	_, err := s.DB.CreateUser(context, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      sql.NullString{String: userName, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error user '%v' is already registered: %v", userName, err)
	}

	// Set the username in the configuration file
	err = s.Cfg.SetUserName(userName)
	if err != nil {
		return fmt.Errorf("error unable to set user name '%v': %v", userName, err)
	}
	fmt.Printf("user '%v' successfully registered\n", userName)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) > 0 {
		return fmt.Errorf("error expects no arguments")
	}

	// Attempt to reset the users database
	// Will remove from all entries from feeds and feed_follows bc cascade
	context := context.Background()
	err := s.DB.ResetUsers(context)
	if err != nil {
		return fmt.Errorf("error unable to reset database: %v", err)
	}
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) > 0 {
		return fmt.Errorf("error expects no arguments")
	}

	context := context.Background()

	// Get the users from the database
	sqlUserNames, err := s.DB.GetUsers(context)
	if err != nil {
		return fmt.Errorf("error unable to get users: %v", err)
	}

	// Output users to the console
	for _, sqlUserName := range sqlUserNames {
		if sqlUserName.Valid {
			fmt.Printf("* %v", sqlUserName.String)
			if sqlUserName.String == s.Cfg.UserName {
				fmt.Print(" (current)")
			}
			fmt.Print("\n")
		}
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) > 0 {
		return fmt.Errorf("error expects no arguments")
	}

	context := context.Background()
	feedURL := "https://www.wagslane.dev/index.xml"
	rssFeed, err := rss.FetchFeed(context, feedURL)
	if err != nil {
		return fmt.Errorf("error unable to fetch feed: %v", err)
	}
	fmt.Printf("%v\n", rssFeed)
	return nil
}

func HandlerAddFeed(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) < 2 {
		return fmt.Errorf("error missing args")
	} else if len(cmd.Args) > 2 {
		return fmt.Errorf("error too many args")
	}

	// Fetch the current user
	context := context.Background()
	name := cmd.Args[0]
	url := cmd.Args[1]
	user, err := s.DB.GetUser(context, sql.NullString{String: s.Cfg.UserName, Valid: true})
	if err != nil {
		return fmt.Errorf("error unable to get current user: %v", err)
	}

	// Attempt to create a feed
	feed, err := s.DB.CreateFeed(context, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      sql.NullString{String: name, Valid: true},
		Url:       sql.NullString{String: url, Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error unable to add entry to feed: %v", err)
	}

	// Attempt to follow the feed
	_, err = s.DB.CreateFeedFollow(context, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error unable to follow the feed: %v", err)
	}

	// Output to console
	fmt.Printf("added feed '%v' (%v) to %v\n", name, url, user.Name.String)
	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) > 0 {
		return fmt.Errorf("error expects no arguments")
	}

	// Get all feeds
	context := context.Background()
	feeds, err := s.DB.GetFeeds(context)
	if err != nil {
		return fmt.Errorf("error unable to get feeds: %v", err)
	}

	// Output the feeds
	for _, feed := range feeds {
		fmt.Printf("* '%v' (%v) - %v\n", feed.Name.String, feed.Url.String, feed.UserName.String)
	}

	return nil
}

func HandlerFollow(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) < 1 {
		return fmt.Errorf("error expects url argument, not none")
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("error expects only one argument")
	}

	context := context.Background()
	feedURL := cmd.Args[0]

	// Fetch the current user
	user, err := s.DB.GetUser(context, sql.NullString{String: s.Cfg.UserName, Valid: true})
	if err != nil {
		return fmt.Errorf("error unable to get user name: %v", err)
	}

	// Fetch the feed
	feed, err := s.DB.GetFeed(context, sql.NullString{String: feedURL, Valid: true})
	if err != nil {
		return fmt.Errorf("error unable to get feed from '%v': %v", feedURL, err)
	}

	// Attempt to follow the feed
	_, err = s.DB.CreateFeedFollow(context, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error unable to add feed follow for '%v' to '%v': %v", user.Name.String, feed.Url.String, err)
	}

	// Output to console
	fmt.Printf(
		"user '%v' is now following '%v'\n",
		user.Name.String,
		feed.Name.String,
	)
	return nil
}

func HandlerFollowing(s *State, cmd Command) error {
	// Validate Args
	if len(cmd.Args) > 0 {
		return fmt.Errorf("error expecting no argument ")
	}

	context := context.Background()

	// Fetch the current user
	user, err := s.DB.GetUser(context, sql.NullString{String: s.Cfg.UserName, Valid: true})
	if err != nil {
		return fmt.Errorf("error unable to get current user %v: %v", s.Cfg.UserName, err)
	}

	// Fetch all feed follows for the current user
	feedFollows, err := s.DB.GetFeedFollowsForUser(context, uuid.NullUUID{UUID: user.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("error unable to get feeds: %v", err)
	}

	// Output to console
	for _, feedFollow := range feedFollows {
		fmt.Printf("* %v\n", feedFollow.FeedName.String)
	}
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
