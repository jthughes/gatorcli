package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jthughes/gatorcli/internal/database"
)

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("command not found")
	}
	return f(s, cmd)
}

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("missing username argument")
	}
	name := cmd.args[0]
	user, err := s.dbq.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("username does not exist")
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Println("User set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("missing username argument")
	}

	user, err := s.dbq.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})
	if err != nil {
		fmt.Printf("unable to create new user: %s\n", err)
		os.Exit(1)
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Println("Created new user:")
	fmt.Printf("%+v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.dbq.ClearUsers(context.Background())
	if err != nil {
		fmt.Println("Unable to clear users from database")
	} else {
		fmt.Println("Cleared users from database")
	}
	return err
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.dbq.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.cfg.Username {
			fmt.Println("*", user.Name, "(current)")
		} else {
			fmt.Println("*", user.Name)
		}
	}
	return nil
}

func handlerAggregator(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("expected one argument: agg <time_between_reqs>")
	}
	duration := cmd.args[0]
	timeBetween, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("unable to convert duration: %w", err)
	}
	fmt.Printf("Collecting feeds every %s\n", duration)

	ticker := time.NewTicker(timeBetween)
	for ; ; <-ticker.C {
		err := scrapeFeeds(context.Background(), s)
		if err != nil {
			return fmt.Errorf("failed to scrape feed: %w", err)
		}
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, loggedInUser database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("wrong number of arguments: expected 'addfeed <name> <url>")
	}
	name, url := cmd.args[0], cmd.args[1]
	feed, err := s.dbq.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    loggedInUser.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to add feed: %w", err)
	}
	fmt.Println("Successfully added feed")
	return middlewareLoggedIn(handlerFollow)(s, command{
		name: "follow",
		args: []string{feed.Url},
	})
}

func handlerGetFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("expected no arguments")
	}
	feeds, err := s.dbq.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("unable to get feeds from database: %w", err)
	}
	for _, feed := range feeds {
		user, err := s.dbq.GetFeedUser(context.Background(), feed.Url)
		if err != nil {
			return fmt.Errorf("unable to find user from feed: %w", err)
		}
		fmt.Printf("* Name: '%s' URL: '%s' Added by: '%s'\n", feed.Name, feed.Url, user)
	}
	return nil
}

func handlerFollow(s *state, cmd command, loggedInUser database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("wrong number of arguments: expected 'follow <url>")
	}
	feedUrl := cmd.args[0]
	feed, err := s.dbq.GetFeedByURL(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("feed url not found: %w", err)
	}
	follow, err := s.dbq.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    loggedInUser.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to create new feed follow: %w", err)
	}
	fmt.Printf("'%s' successfully followed '%s' feed\n", follow.UserName, follow.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("expected no arguments, received %d", len(cmd.args))
	}
	follows, err := s.dbq.GetFeedFollowsForUser(context.Background(), s.cfg.Username)
	if err != nil {
		return fmt.Errorf("current user follows not found: %w", err)
	}
	fmt.Printf("%s's feeds:\n", s.cfg.Username)
	for _, feed := range follows {
		fmt.Printf("* %s\n", feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, loggedInUser database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("expected one argument: unfollow <feed_url>")
	}
	url := cmd.args[0]
	err := s.dbq.UnfollowFeed(context.Background(), database.UnfollowFeedParams{
		Name: loggedInUser.Name,
		Url:  url,
	})
	if err != nil {
		return fmt.Errorf("unable to unfollow feed: %w", err)
	}
	fmt.Printf("%s unfollowed feed at '%s'\n", loggedInUser.Name, url)
	return nil
}
