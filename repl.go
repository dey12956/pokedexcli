package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dey12956/pokedexcli/internal/pokeapi"
	"github.com/peterh/liner"
)

func startRepl(c *config) {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetCompleter(func(lineText string) []string {
		matches := make([]string, 0)
		for name := range getCommands() {
			if strings.HasPrefix(name, lineText) {
				matches = append(matches, name)
			}
		}
		sort.Strings(matches)
		return matches
	})

	for {
		input, err := line.Prompt("Pokedex > ")
		if err != nil {
			if errors.Is(err, liner.ErrPromptAborted) {
				fmt.Println()
				continue
			}
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		line.AppendHistory(input)
		words := cleanInput(input)
		if len(words) == 0 {
			continue
		}

		command, exists := getCommands()[words[0]]

		if exists {
			if len(words) > 1 {
				err := command.callback(c, words[1:]...)
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
			description: "Get the next page of locations (number to explore, arrows to page)",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Get the previous page of locations (number to explore, arrows to page)",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "Get Pokemon located in the specified area (enter a number to battle)",
			callback:    commandExplore,
		},
		"battle": {
			name:        "battle",
			description: "Battle a wild Pokemon",
			callback:    commandBattle,
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
	pokeapiClient  pokeapi.Client
	Next           *string
	Previous       *string
	mapFetched     bool
	Pokedex        map[string][]Pokemon
	Inventory      Inventory
	UserName       string
	StoragePath    string
	LastDailyGrant string
}

type pokemonAbility struct {
	name     string
	isHidden bool
	slot     int
}

type PokemonMove struct {
	name     string
	power    int
	accuracy int
	moveType string
	priority int
}

type Inventory struct {
	Pokeball  int
	GreatBall int
	UltraBall int
	Potion    int
}

type Pokemon struct {
	name           string
	dateCaught     time.Time
	height         int
	weight         int
	stats          map[string]int
	types          []string
	id             int
	baseExperience int
	order          int
	isDefault      bool
	species        string
	abilities      []pokemonAbility
	heldItems      []string
	forms          []string
	moves          []PokemonMove
	moveCount      int
	level          int
	experience     int
	growthRate     string
	evolutionChain string
	lastXPAt       time.Time
	lastXPGain     int
}
