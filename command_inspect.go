package main

import (
	"errors"
	"fmt"
)

func commandInspect(c *config, name ...string) error {
	if len(name) == 0 {
		return errors.New("Enter a Pokemon to inspect")
	}
	if len(name) > 1 {
		return errors.New("Command inspect takes a single Pokemon")
	}

	fmt.Println()

	if poke, exists := c.Pokedex[name[0]]; exists {
		fmt.Printf("Name: %s\n", poke.name)
		fmt.Printf("Height: %v\n", poke.height)
		fmt.Printf("Weight: %v\n", poke.weight)
		fmt.Println("Stats:")
		fmt.Printf("-hp: %v\n", poke.stats["hp"])
		fmt.Printf("-attack: %v\n", poke.stats["attack"])
		fmt.Printf("-defense: %v\n", poke.stats["defense"])
		fmt.Printf("-special-attack: %v\n", poke.stats["special-attack"])
		fmt.Printf("-special-defense: %v\n", poke.stats["special-defense"])
		fmt.Printf("-speed: %v\n", poke.stats["speed"])
		fmt.Println("Types:")
		for _, poketype := range poke.types {
			fmt.Printf("-%s\n", poketype)
		}
	} else {
		fmt.Println("You have not caught that pokemon")
	}

	fmt.Println()

	return nil
}
