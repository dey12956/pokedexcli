package main

import "math/rand"

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

	pokemon, err := buildPokemonFromResponse(c, catchPokeResp)
	if err != nil {
		return false, err
	}
	appendCaughtPokemon(c, pokemon)

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
