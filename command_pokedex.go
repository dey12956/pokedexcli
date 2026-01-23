package main

import (
	"errors"
	"fmt"
)

func commandPokedex(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command pokedex doesn't take arguments")
	}

	fmt.Println()
	fmt.Println("Your Pokedex:")
	for name := range c.Pokedex {
		fmt.Printf("-%s\n", name)
	}
	fmt.Println()

	return nil
}
