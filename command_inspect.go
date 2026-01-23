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

	if entries, exists := c.Pokedex[name[0]]; exists && len(entries) > 0 {
		poke := entries[len(entries)-1]
		fmt.Printf("Name: %s\n", poke.name)
		fmt.Printf("ID: %d\n", poke.id)
		if poke.species != "" {
			fmt.Printf("Species: %s\n", poke.species)
		}
		if poke.level > 0 {
			fmt.Printf("Level: %d\n", poke.level)
		}
		if poke.experience > 0 {
			fmt.Printf("XP: %d\n", poke.experience)
		}
		fmt.Printf("Base XP: %d\n", poke.baseExperience)
		fmt.Printf("Height: %v\n", poke.height)
		fmt.Printf("Weight: %v\n", poke.weight)
		fmt.Printf("Order: %d\n", poke.order)
		fmt.Printf("Default: %t\n", poke.isDefault)
		fmt.Printf("Moves: %d\n", poke.moveCount)
		if len(poke.moves) > 0 {
			fmt.Println("Move details:")
			for _, move := range poke.moves {
				fmt.Printf("-%s (power %d, accuracy %d, priority %d, type %s)\n", move.name, move.power, move.accuracy, move.priority, move.moveType)
			}
		}
		fmt.Println("Abilities:")
		if len(poke.abilities) == 0 {
			fmt.Println("-none")
		} else {
			for _, ability := range poke.abilities {
				label := fmt.Sprintf("-%s (slot %d)", ability.name, ability.slot)
				if ability.isHidden {
					label = fmt.Sprintf("%s hidden", label)
				}
				fmt.Println(label)
			}
		}
		fmt.Println("Held items:")
		if len(poke.heldItems) == 0 {
			fmt.Println("-none")
		} else {
			for _, item := range poke.heldItems {
				fmt.Printf("-%s\n", item)
			}
		}
		if len(poke.forms) > 0 {
			fmt.Println("Forms:")
			for _, form := range poke.forms {
				fmt.Printf("-%s\n", form)
			}
		}
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
