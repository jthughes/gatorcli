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
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return nil
	}
	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("wrong number of arguments: expected 'addfeed <name> <url>")
	}
	name, url := cmd.args[0], cmd.args[1]
	user, err := s.dbq.GetUser(context.Background(), s.cfg.Username)
	if err != nil {
		return fmt.Errorf("user not found in database: %w", err)
	}
	feed, err := s.dbq.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to add feed: %w", err)
	}
	fmt.Println("Successfully added feed")
	fmt.Printf("%+v\n", feed)
	return nil
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
