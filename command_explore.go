package main

import (
	"errors"
	"fmt"
)

func commandExplore(c *config, name ...string) error {
	if len(name) == 0 {
		return errors.New("Enter an area to explore")
	}

	fmt.Println()
	fmt.Printf("Exploring %s...\n", name[0])
	fmt.Println("Found Pokemon:")

	pokemonResp, err := c.pokeapiClient.ListPokemon(name[0])
	if err != nil {
		return err
	}

	for _, pokemonEncounter := range pokemonResp.PokemonEncounters {
		fmt.Println("-" + pokemonEncounter.Pokemon.Name)
	}

	fmt.Println()

	return nil
}
