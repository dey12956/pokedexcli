package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ensureStarterPokemon(c *config, dataExists bool) error {
	if c == nil {
		return nil
	}
	if dataExists && len(c.Pokedex) > 0 {
		return nil
	}
	choices := []string{"bulbasaur", "charmander", "squirtle"}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Choose your starter Pokemon:")
		for i, name := range choices {
			fmt.Printf("%d) %s\n", i+1, starterDisplayName(name))
		}
		fmt.Print("Starter # > ")
		input, needsNewline, err := readLine(reader)
		if err != nil {
			return err
		}
		if needsNewline {
			fmt.Println()
		}
		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("Enter 1-3 to choose your starter.")
			continue
		}
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(choices) {
			fmt.Println("Enter 1-3 to choose your starter.")
			continue
		}
		name := choices[choice-1]
		resp, err := c.pokeapiClient.GetPokemon(name)
		if err != nil {
			return err
		}
		pokemon, err := buildPokemonFromResponse(c, resp)
		if err != nil {
			return err
		}
		appendCaughtPokemon(c, pokemon)
		if err := saveUserData(c); err != nil {
			return err
		}
		fmt.Printf("Starter %s added to your Pokedex.\n", starterDisplayName(name))
		return nil
	}
}

func starterDisplayName(name string) string {
	if name == "" {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
