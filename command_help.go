package main

import "fmt"

func commandHelp(c *config) error {
	fmt.Println()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for name, command := range getCommands() {
		fmt.Printf("%s: %s\n", name, command.description)
	}
	fmt.Println()
	return nil
}

