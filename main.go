package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jthughes/gatorcli/internal/config"
	"github.com/jthughes/gatorcli/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	dbq *database.Queries
}

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

func main() {
	cfg := config.Read()
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		fmt.Println("unable to open connection to database: ", err)
		os.Exit(1)
	}

	programState := state{
		cfg: &cfg,
		dbq: database.New(db),
	}
	cmds := commands{
		handlers: map[string]func(*state, command) error{},
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Require an argument, received", len(args)-1)
		os.Exit(1)
	}
	err = cmds.run(&programState, command{name: args[1], args: args[2:]})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
