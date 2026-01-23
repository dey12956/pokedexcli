package main

import (
	"errors"
	"fmt"
	"sort"
)

func commandHelp(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command help doesn't take arguments")
	}

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
