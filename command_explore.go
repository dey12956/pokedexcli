package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func commandExplore(c *config, name ...string) error {
	if len(name) == 0 {
		return errors.New("Enter an area to explore")
	}
	if len(name) > 1 {
		return errors.New("Command explore takes a single area")
	}

	fmt.Println()
	fmt.Printf("Exploring %s...\n", name[0])
	fmt.Println("Found Pokemon:")

	pokemonResp, err := c.pokeapiClient.ListPokemon(name[0])
	if err != nil {
		return err
	}

	if len(pokemonResp.PokemonEncounters) == 0 {
		fmt.Println()
		return nil
	}

	pokemonNames := make([]string, 0, len(pokemonResp.PokemonEncounters))
	for i, pokemonEncounter := range pokemonResp.PokemonEncounters {
		pokemonNames = append(pokemonNames, pokemonEncounter.Pokemon.Name)
		fmt.Printf("%d) %s\n", i+1, pokemonEncounter.Pokemon.Name)
	}
	fmt.Println("Hint: enter a number to battle a Pokemon from this area.")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Battle # (or press Enter to continue) > ")
		input, needsNewline, err := readLine(reader)
		if err != nil {
			return err
		}
		if needsNewline {
			fmt.Println()
		}
		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println()
			return nil
		}
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(pokemonNames) {
			fmt.Printf("Enter 1-%d, or press Enter to continue.\n", len(pokemonNames))
			continue
		}
		fmt.Println()
		return commandBattle(c, pokemonNames[choice-1])
	}
}
