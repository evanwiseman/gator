package main

import (
	"fmt"

	"github.com/evanwiseman/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	cfg.SetUsername("evan")

	new_cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Printf("%v\n", new_cfg)
}
