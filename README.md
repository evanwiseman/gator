# gator

## Introduction
gator aggregates RSS Feeds and allows users to follow and add any RSS Feed to their currated list of RSS Feeds.

# Requirements
Go
Postgres


## Install
Download the repository and install with ```go install .``` this will add it to your $GOBIN or ~/go/bin. You should be able to use the CLI tool with "gator" inside your terminal.

Create a .gatorconfig.json in your $HOME directory (~) with the key "db_url". db_url will point to a postgres database locally configured on your machine.

## Commands
- "login" (usage: "login <name>"): Allows a user to login to their account and access their feeds.
- "register" (usage: "register <name>"): Registers a user with that name in the database.
- "reset" (usage: "reset"): Resets the user database.
- "users" (usage: "users"): Lists all users in the database.
- "addfeed" (usage: "addfeed <name> <url>"): Adds a feed to the users profile with the given name and url.
- "feeds" (usage: "feeds"): Lists all feeds in the database.
- "follow" (usage: "follow <url>"): Follows a RSS Feed by providing the URL to the feed.
- "unfollow" (usage: "unfollow <url>"): Unfollows a RSS feed by providing the URL.
- "following" (usage: "following"): Provides a list of feeds the current user is following.
- "agg" (usage: "agg <time_duration>): Aggregates posts from the feeds the current user is following. Set a time duration as (1s, 1m, 1h).
- "browse" (usage: "browse [limit]): Grabs the most recent posts aggregated in the database for the user. limit defaults to 2.