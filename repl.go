package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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

		command, exists := getCommands()[words[0]]

		if exists {
			if len(words) > 1 {
				err := command.callback(c, words[1])
				if err != nil {
					if errors.Is(err, errExit) {
						return
					}
					fmt.Println(err)
				}
				continue
			}
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
	callback    func(*config, ...string) error
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
			description: "Display a help message",
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
		"explore": {
			name:        "explore",
			description: "Get Pokemon located in the specified area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Throw a Pokeball at the specified Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a Pokemon you have caught before",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Show your Pokedex",
			callback:    commandPokedex,
		},
		"tui": {
			name:        "tui",
			description: "Launch the TUI map explorer",
			callback:    commandTui,
		},
	}
}

type config struct {
	pokeapiClient pokeapi.Client
	Next          *string
	Previous      *string
	mapFetched    bool
	Pokedex       map[string]Pokemon
}

type Pokemon struct {
	name       string
	dateCaught time.Time
	height     int
	weight     int
	stats      map[string]int
	types      []string
}
