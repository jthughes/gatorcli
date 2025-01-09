package main

import (
	"context"
	"fmt"

	"github.com/jthughes/gatorcli/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.dbq.GetUser(context.Background(), s.cfg.Username)
		if err != nil {
			return fmt.Errorf("logged in user not found: %w", err)
		}
		return handler(s, cmd, user)
	}
}
