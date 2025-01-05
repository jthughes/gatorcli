package main

import (
	"fmt"

	"github.com/jthughes/gatorcli/internal/config"
)

func main() {
	cfg := config.Read()
	cfg.SetUser("test")
	cfg = config.Read()
	fmt.Println(cfg)
}
