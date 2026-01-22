package main

import (
	"bufio"
	"os"
	"fmt"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		register()
		if !scanner.Scan() { break }
		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 0 { continue }
		if command, ok := commands[words[0]]; ok{
			err := command.callback()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

type cliCommand struct {
	name string
	description string
	callback func() error
}

var commands map[string]cliCommand

func register() {
	commands = map[string]cliCommand{
		"exit": {
			name: "exit",
			description: "Exit the Pokedex",
			callback: commandExit,
		},
		"help": {
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
	}
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for name, command := range commands {
		fmt.Printf("%s: %s\n", name, command.description)
	}
	return nil
}
