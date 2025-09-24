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

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		context := context.Background()
		user, err := s.DB.GetUser(context, sql.NullString{
			String: s.Cfg.UserName,
			Valid:  true,
		})
		if err != nil {
			return fmt.Errorf("must be logged in: %v", err)
		}
		return handler(s, cmd, user)
	}
}

func HandlerLogin(s *State, cmd Command) error {
	// Validate Args
	usage := "usage: login <name>"
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username provided. %v", usage)
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("more than one username provided. %v", usage)
	}

	// Check user is in database
	context := context.Background()
	userName := cmd.Args[0]
	_, err := s.DB.GetUser(context, sql.NullString{String: userName, Valid: true})
	if err != nil {
		return fmt.Errorf("user not in database: %v", err)
	}

	// Set the user in the database
	err = s.Cfg.SetUserName(userName)
	if err != nil {
		return fmt.Errorf("unable to set username: %v", err)
	}
	fmt.Printf("user successfully set to %v\n", userName)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	// Validate Args
	usage := "usage: register <name>"
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username provided. %v", usage)
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("more than one username provided. %v", usage)
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
		return fmt.Errorf("user '%v' is already registered: %v", userName, err)
	}

	// Set the username in the configuration file
	err = s.Cfg.SetUserName(userName)
	if err != nil {
		return fmt.Errorf("unable to set user name '%v': %v", userName, err)
	}
	fmt.Printf("user '%v' successfully registered\n", userName)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	// Validate Args
	usage := "usage: reset"
	if len(cmd.Args) > 0 {
		return fmt.Errorf("no arguments required for reset. %v", usage)
	}

	// Attempt to reset the users database
	// Will remove from all entries from feeds and feed_follows bc cascade
	context := context.Background()
	err := s.DB.ResetUsers(context)
	if err != nil {
		return fmt.Errorf("unable to reset database: %v", err)
	}
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	// Validate Args
	usage := "usage: users"
	if len(cmd.Args) > 0 {
		return fmt.Errorf("no arguments required to list users. %v", usage)
	}

	context := context.Background()

	// Get the users from the database
	sqlUserNames, err := s.DB.GetUsers(context)
	if err != nil {
		return fmt.Errorf("unable to get users: %v", err)
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
		return fmt.Errorf("unable to fetch feed: %v", err)
	}
	fmt.Printf("%v\n", rssFeed)
	return nil
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	// Validate Args
	usage := "usage: addfeed <name> <url>"
	if len(cmd.Args) < 2 {
		return fmt.Errorf("missing argument(s). %v", usage)
	} else if len(cmd.Args) > 2 {
		return fmt.Errorf("too many argument(s). %v", usage)
	}

	// Fetch the current user
	context := context.Background()
	name := cmd.Args[0]
	url := cmd.Args[1]

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
		return fmt.Errorf("unable to add entry to feed: %v", err)
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
		return fmt.Errorf("unable to follow the feed: %v", err)
	}

	// Output to console
	fmt.Printf("added feed '%v' (%v) to %v\n", name, url, user.Name.String)
	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	// Validate Args
	usage := "usage: feeds"
	if len(cmd.Args) > 0 {
		return fmt.Errorf("no arguments required to list feeds. %v", usage)
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

func HandlerFollow(s *State, cmd Command, user database.User) error {
	// Validate Args
	usage := "usage: follow <url>"
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing url. %v", usage)
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("more than one url provided. %v", usage)
	}

	context := context.Background()
	feedURL := cmd.Args[0]

	// Fetch the feed
	feed, err := s.DB.GetFeed(context, sql.NullString{String: feedURL, Valid: true})
	if err != nil {
		return fmt.Errorf("unable to get feed from '%v': %v", feedURL, err)
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
		return fmt.Errorf("unable to add feed follow for '%v' to '%v': %v", user.Name.String, feed.Url.String, err)
	}

	// Output to console
	fmt.Printf(
		"user '%v' is now following '%v'\n",
		user.Name.String,
		feed.Name.String,
	)
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	// Validate Args
	usage := "usage: unfollow <url>"
	if len(cmd.Args) < 1 {
		return fmt.Errorf("missing url. %v", usage)
	} else if len(cmd.Args) > 1 {
		return fmt.Errorf("more than one url provided. %v", usage)
	}

	context := context.Background()
	url := cmd.Args[0]

	err := s.DB.DeleteFeedFollow(context, database.DeleteFeedFollowParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		Url:    sql.NullString{String: url, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("unable to unfollow '%v': %v", url, err)
	}
	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	// Validate Args
	usage := "usage: following"
	if len(cmd.Args) > 0 {
		return fmt.Errorf("no arguments required to list followings. %v", usage)
	}

	context := context.Background()

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
