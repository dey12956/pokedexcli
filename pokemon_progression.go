package main

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

const maxLevel = 100

func levelForExperience(c *config, growthRateURL string, experience int) (int, error) {
	if strings.TrimSpace(growthRateURL) == "" {
		return 1, nil
	}
	if experience < 0 {
		experience = 0
	}
	resp, err := c.pokeapiClient.GetGrowthRate(growthRateURL)
	if err != nil {
		return 1, err
	}
	level := 1
	for _, entry := range resp.Levels {
		if entry.Experience <= experience {
			level = entry.Level
		}
	}
	if level > maxLevel {
		level = maxLevel
	}
	return level, nil
}

func applyRestXP(c *config, pokemon *Pokemon) error {
	if pokemon == nil || pokemon.lastXPAt.IsZero() || pokemon.lastXPGain <= 0 {
		return nil
	}
	elapsed := time.Since(pokemon.lastXPAt).Hours()
	if elapsed <= 0 {
		return nil
	}
	xpPerHour := int(math.Round(float64(pokemon.lastXPGain) * 0.05))
	if xpPerHour <= 0 {
		return nil
	}
	bonus := int(float64(xpPerHour) * elapsed)
	cap := pokemon.lastXPGain * 2
	if bonus > cap {
		bonus = cap
	}
	if bonus <= 0 {
		return nil
	}

	if err := applyExperience(c, pokemon, bonus, false); err != nil {
		return err
	}
	pokemon.lastXPAt = time.Now()
	return nil
}

func applyExperience(c *config, pokemon *Pokemon, gained int, updateLastGain bool) error {
	if pokemon == nil || gained <= 0 {
		return nil
	}

	if pokemon.level >= maxLevel {
		pokemon.level = maxLevel
		return nil
	}

	pokemon.experience += gained
	level, err := levelForExperience(c, pokemon.growthRate, pokemon.experience)
	if err != nil {
		return err
	}
	if level > maxLevel {
		level = maxLevel
	}
	prevLevel := pokemon.level
	pokemon.level = level
	if updateLastGain {
		pokemon.lastXPGain = gained
		pokemon.lastXPAt = time.Now()
	}
	if pokemon.level > prevLevel {
		if err := maybeEvolve(c, pokemon); err != nil {
			return err
		}
		grantRandomSupplies(c, "Level up")
		return nil
	}
	return nil
}

func maybeEvolve(c *config, pokemon *Pokemon) error {
	if pokemon == nil || pokemon.evolutionChain == "" || pokemon.species == "" {
		return nil
	}
	chain, err := c.pokeapiClient.GetEvolutionChain(pokemon.evolutionChain)
	if err != nil {
		return err
	}

	nextName, minLevel, found := findNextEvolution(chain.Chain, pokemon.species)
	if !found || minLevel == 0 || pokemon.level < minLevel {
		return nil
	}

	resp, err := c.pokeapiClient.GetPokemon(nextName)
	if err != nil {
		return err
	}
	updated, err := buildPokemonFromResponse(c, resp)
	if err != nil {
		return err
	}

	updated.experience = pokemon.experience
	updated.level = pokemon.level
	updated.growthRate = pokemon.growthRate
	updated.evolutionChain = pokemon.evolutionChain
	updated.lastXPAt = pokemon.lastXPAt
	updated.lastXPGain = pokemon.lastXPGain
	updated.dateCaught = pokemon.dateCaught

	*pokemon = updated
	return nil
}

func findNextEvolution(link pokeapi.EvolutionChainLink, species string) (string, int, bool) {
	if strings.EqualFold(link.Species.Name, species) {
		minLevel := 0
		var nextName string
		for _, evolve := range link.EvolvesTo {
			candidateLevel := 0
			for _, detail := range evolve.EvolutionDetails {
				if detail.MinLevel != nil && *detail.MinLevel > candidateLevel {
					candidateLevel = *detail.MinLevel
				}
			}
			if nextName == "" || (candidateLevel > 0 && candidateLevel < minLevel) || minLevel == 0 {
				minLevel = candidateLevel
				nextName = evolve.Species.Name
			}
		}
		if nextName != "" {
			return nextName, minLevel, true
		}
	}
	for _, evolve := range link.EvolvesTo {
		if nextName, minLevel, found := findNextEvolution(evolve, species); found {
			return nextName, minLevel, true
		}
	}
	return "", 0, false
}

var errNoPokemon = errors.New("no pokemon available")
