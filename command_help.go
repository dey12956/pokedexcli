package main

import (
	"fmt"
	"sort"
)

func commandHelp(c *config, name ...string) error {
	fmt.Println()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	commands := getCommands()
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("%s: %s\n", name, commands[name].description)
	}
	fmt.Println()
	return nil
}
