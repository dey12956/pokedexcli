package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func commandCatch(c *config, name ...string) error {
	if len(name) == 0 {
		return errors.New("Enter an Pokemon to catch")
	}
	if len(name) > 1 {
		return errors.New("Command catch takes a single Pokemon")
	}

	fmt.Println()
	fmt.Printf("Throwing a Pokeball at %s...\n", name[0])

	caught, err := attemptCatch(c, name[0])
	if err != nil {
		return err
	}
	if caught {
		fmt.Printf("%s was caught!\n", name[0])
		if err := saveUserData(c); err != nil {
			fmt.Printf("Warning: failed to save data: %v\n", err)
		}
	} else {
		fmt.Printf("%s escaped!\n", name[0])
		return nil
	}

	fmt.Println()

	return nil
}

func attemptCatch(c *config, name string) (bool, error) {
	catchPokeResp, err := c.pokeapiClient.CatchPokemon(name)
	if err != nil {
		return false, err
	}

	baseXP := catchPokeResp.BaseExperience
	p := catchProb(baseXP)

	if rand.Float64() >= p {
		return false, nil
	}

	stats := make(map[string]int)
	for _, stat := range catchPokeResp.Stats {
		stats[stat.Stat.Name] = stat.BaseStat
	}

	types := make([]string, 0, len(catchPokeResp.Types))
	for _, poketype := range catchPokeResp.Types {
		types = append(types, poketype.Type.Name)
	}

	abilities := make([]pokemonAbility, 0, len(catchPokeResp.Abilities))
	for _, ability := range catchPokeResp.Abilities {
		abilities = append(abilities, pokemonAbility{
			name:     ability.Ability.Name,
			isHidden: ability.IsHidden,
			slot:     ability.Slot,
		})
	}

	heldItems := make([]string, 0, len(catchPokeResp.HeldItems))
	for _, item := range catchPokeResp.HeldItems {
		heldItems = append(heldItems, item.Item.Name)
	}

	forms := make([]string, 0, len(catchPokeResp.Forms))
	for _, form := range catchPokeResp.Forms {
		forms = append(forms, form.Name)
	}

	c.Pokedex[name] = Pokemon{
		name:           name,
		dateCaught:     time.Now(),
		height:         catchPokeResp.Height,
		weight:         catchPokeResp.Weight,
		stats:          stats,
		types:          types,
		id:             catchPokeResp.ID,
		baseExperience: catchPokeResp.BaseExperience,
		order:          catchPokeResp.Order,
		isDefault:      catchPokeResp.IsDefault,
		species:        catchPokeResp.Species.Name,
		abilities:      abilities,
		heldItems:      heldItems,
		forms:          forms,
		moveCount:      len(catchPokeResp.Moves),
	}

	return true, nil
}

func catchProb(baseXP int) float64 {
	const (
		minXP = 36
		kink  = 255
		maxXP = 608

		pEasy = 0.79
		pK    = 0.30
		pHard = 0.08
	)

	if baseXP <= minXP {
		return pEasy
	}
	if baseXP >= maxXP {
		return pHard
	}

	x := float64(baseXP)

	if baseXP <= kink {
		t := (x - minXP) / float64(kink-minXP)
		return pEasy + (pK-pEasy)*t
	}

	t := (x - kink) / float64(maxXP-kink)
	return pK + (pHard-pK)*t
}
