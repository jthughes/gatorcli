package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jthughes/gatorcli/internal/config"
	"github.com/jthughes/gatorcli/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	dbq *database.Queries
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
	cmds.register("agg", handlerAggregator)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerGetFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)
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
