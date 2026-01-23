package main

import (
	"errors"
	"fmt"
	"sort"
)

func commandPokedex(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command pokedex doesn't take arguments")
	}

	fmt.Println()
	fmt.Println("Your Pokedex:")
	if len(c.Pokedex) == 0 {
		fmt.Println("-empty")
		fmt.Println()
		return nil
	}
	names := make([]string, 0, len(c.Pokedex))
	for name := range c.Pokedex {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("-%s (x%d)\n", name, len(c.Pokedex[name]))
	}
	fmt.Println()

	return nil
}
