package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

func startRepl(c *config) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 0 {
			continue
		}

		if command, exists := getCommands()[words[0]]; exists {
			err := command.callback(c)
			if err != nil {
				if errors.Is(err, errExit) {
					return
				}
				fmt.Println(err)
			}
			continue
		} else {
			fmt.Println("Unknown command")
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Input error:", err)
	}
}

func cleanInput(text string) []string {
	lowerCaseString := strings.ToLower(text)
	return strings.Fields(lowerCaseString)
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Get the next page of locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Get the previous page of locations",
			callback:    commandMapB,
		},
	}
}

type config struct {
	pokeapiClient pokeapi.Client
	Next          *string
	Previous      *string
	mapFetched    bool
}
