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

	fmt.Println()
	fmt.Printf("Throwing a Pokeball at %s...\n", name[0])

	catchPokeResp, err := c.pokeapiClient.CatchPokemon(name[0])
	if err != nil {
		return err
	}

	baseXP := catchPokeResp.BaseExperience
	p := catchProb(baseXP)

	if rand.Float64() < p {
		fmt.Printf("%s escaped!\n", name[0])
		return nil
	} else {
		fmt.Printf("%s was caught!\n", name[0])
	}

	fmt.Println()

	stats := make(map[string]int)
	for _, stat := range catchPokeResp.Stats {
		stats[stat.Stat.Name] = stat.BaseStat
	}

	types := make([]string, 0, len(catchPokeResp.Types))
	for _, poketype := range catchPokeResp.Types {
		types = append(types, poketype.Type.Name)
	}

	c.Pokedex[name[0]] = Pokemon{
		name:       name[0],
		dateCaught: time.Now(),
		height:     catchPokeResp.Height,
		weight:     catchPokeResp.Weight,
		stats:      stats,
		types:      types,
	}

	return nil
}

func catchProb(baseXP int) float64 {
	const (
		minXP = 36
		kink  = 255
		maxXP = 608

		pMin = 0.08
		pK   = 0.40
		pMax = 0.95
	)

	if baseXP <= minXP {
		return pMin
	}
	if baseXP >= maxXP {
		return pMax
	}

	x := float64(baseXP)

	if baseXP <= kink {
		t := (x - minXP) / float64(kink-minXP)
		return pMin + (pK-pMin)*t
	}

	t := (x - kink) / float64(maxXP-kink)
	return pK + (pMax-pK)*t
}
