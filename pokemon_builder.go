package main

import (
	"time"

	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

func buildPokemonFromResponse(c *config, resp pokeapi.CatchPokemonResponse) (Pokemon, error) {
	stats := make(map[string]int)
	for _, stat := range resp.Stats {
		stats[stat.Stat.Name] = stat.BaseStat
	}

	types := make([]string, 0, len(resp.Types))
	for _, poketype := range resp.Types {
		types = append(types, poketype.Type.Name)
	}

	abilities := make([]pokemonAbility, 0, len(resp.Abilities))
	for _, ability := range resp.Abilities {
		abilities = append(abilities, pokemonAbility{
			name:     ability.Ability.Name,
			isHidden: ability.IsHidden,
			slot:     ability.Slot,
		})
	}

	heldItems := make([]string, 0, len(resp.HeldItems))
	for _, item := range resp.HeldItems {
		heldItems = append(heldItems, item.Item.Name)
	}

	forms := make([]string, 0, len(resp.Forms))
	for _, form := range resp.Forms {
		forms = append(forms, form.Name)
	}

	moves := make([]PokemonMove, 0, len(resp.Moves))
	for _, move := range resp.Moves {
		moveResp, err := c.pokeapiClient.GetMove(move.Move.URL)
		if err != nil {
			return Pokemon{}, err
		}
		power := 0
		if moveResp.Power != nil {
			power = *moveResp.Power
		}
		accuracy := 0
		if moveResp.Accuracy != nil {
			accuracy = *moveResp.Accuracy
		}
		moves = append(moves, PokemonMove{
			name:     moveResp.Name,
			power:    power,
			accuracy: accuracy,
			priority: moveResp.Priority,
			moveType: moveResp.Type.Name,
		})
	}

	speciesResp, err := c.pokeapiClient.GetPokemonSpecies(resp.Species.Name)
	if err != nil {
		return Pokemon{}, err
	}

	level, err := levelForExperience(c, speciesResp.GrowthRate.URL, resp.BaseExperience)
	if err != nil {
		return Pokemon{}, err
	}

	return Pokemon{
		name:           resp.Name,
		dateCaught:     time.Now(),
		height:         resp.Height,
		weight:         resp.Weight,
		stats:          stats,
		types:          types,
		id:             resp.ID,
		baseExperience: resp.BaseExperience,
		order:          resp.Order,
		isDefault:      resp.IsDefault,
		species:        resp.Species.Name,
		abilities:      abilities,
		heldItems:      heldItems,
		forms:          forms,
		moves:          moves,
		moveCount:      len(resp.Moves),
		level:          level,
		experience:     resp.BaseExperience,
		growthRate:     speciesResp.GrowthRate.URL,
		evolutionChain: speciesResp.EvolutionChain.URL,
		lastXPAt:       time.Now(),
		lastXPGain:     resp.BaseExperience,
	}, nil
}
