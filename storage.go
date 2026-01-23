package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/peterh/liner"
)

type userData struct {
	User           string                     `json:"user"`
	Pokedex        map[string][]pokemonRecord `json:"pokedex"`
	Inventory      inventoryRecord            `json:"inventory"`
	LastDailyGrant string                     `json:"last_daily_grant"`
}

type inventoryRecord struct {
	Pokeball  int `json:"pokeball"`
	GreatBall int `json:"great_ball"`
	UltraBall int `json:"ultra_ball"`
	Potion    int `json:"potion"`
}

type pokemonAbilityRecord struct {
	Name     string `json:"name"`
	IsHidden bool   `json:"is_hidden"`
	Slot     int    `json:"slot"`
}

type pokemonRecord struct {
	Name           string                 `json:"name"`
	DateCaught     time.Time              `json:"date_caught"`
	Height         int                    `json:"height"`
	Weight         int                    `json:"weight"`
	Stats          map[string]int         `json:"stats"`
	Types          []string               `json:"types"`
	ID             int                    `json:"id"`
	BaseExperience int                    `json:"base_experience"`
	Order          int                    `json:"order"`
	IsDefault      bool                   `json:"is_default"`
	Species        string                 `json:"species"`
	Abilities      []pokemonAbilityRecord `json:"abilities"`
	HeldItems      []string               `json:"held_items"`
	Forms          []string               `json:"forms"`
	Moves          []pokemonMoveRecord    `json:"moves"`
	MoveCount      int                    `json:"move_count"`
	Level          int                    `json:"level"`
	Experience     int                    `json:"experience"`
	GrowthRate     string                 `json:"growth_rate"`
	EvolutionChain string                 `json:"evolution_chain"`
	LastXPAt       time.Time              `json:"last_xp_at"`
	LastXPGain     int                    `json:"last_xp_gain"`
}

type pokemonMoveRecord struct {
	Name     string `json:"name"`
	Power    int    `json:"power"`
	Accuracy int    `json:"accuracy"`
	Type     string `json:"type"`
	Priority int    `json:"priority"`
}

func promptUserName() (string, error) {
	defaultName := defaultUserName()
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	input, err := line.Prompt("Trainer name: ")
	if err != nil {
		if errors.Is(err, liner.ErrPromptAborted) || errors.Is(err, io.EOF) {
			return defaultName, nil
		}
		return "", err
	}
	name := strings.TrimSpace(input)
	if name == "" {
		name = defaultName
	}
	return name, nil
}

func defaultUserName() string {
	for _, key := range []string{"USER", "LOGNAME", "USERNAME"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return "trainer"
}

func sanitizeUserName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = strings.ToLower(name)
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ' || r == '.':
			b.WriteRune('_')
		default:
			b.WriteRune('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func userDataPath(userName string) (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(baseDir, "pokedexcli")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", err
	}
	fileName := sanitizeUserName(userName)
	if fileName == "" {
		fileName = "trainer"
	}
	return filepath.Join(dataDir, fileName+".json"), nil
}

func loadUserData(path string) (map[string][]Pokemon, Inventory, string, error) {
	if path == "" {
		return make(map[string][]Pokemon), defaultInventory(), "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]Pokemon), defaultInventory(), "", nil
		}
		return nil, Inventory{}, "", err
	}

	var raw struct {
		User           string                     `json:"user"`
		Pokedex        map[string]json.RawMessage `json:"pokedex"`
		Inventory      *inventoryRecord           `json:"inventory"`
		LastDailyGrant string                     `json:"last_daily_grant"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, Inventory{}, "", err
	}
	result := make(map[string][]Pokemon, len(raw.Pokedex))
	for name, payload := range raw.Pokedex {
		var records []pokemonRecord
		if err := json.Unmarshal(payload, &records); err == nil {
			pokemons := make([]Pokemon, 0, len(records))
			for _, record := range records {
				pokemons = append(pokemons, recordToPokemon(record))
			}
			result[name] = pokemons
			continue
		}
		var record pokemonRecord
		if err := json.Unmarshal(payload, &record); err != nil {
			return nil, Inventory{}, "", err
		}
		result[name] = []Pokemon{recordToPokemon(record)}
	}
	inv := defaultInventory()
	if raw.Inventory != nil {
		inv = inventoryFromRecord(*raw.Inventory)
	}
	return result, inv, raw.LastDailyGrant, nil
}

func saveUserData(c *config) error {
	if c == nil || c.StoragePath == "" {
		return nil
	}
	records := make(map[string][]pokemonRecord, len(c.Pokedex))
	for name, pokes := range c.Pokedex {
		entries := make([]pokemonRecord, 0, len(pokes))
		for _, poke := range pokes {
			entries = append(entries, pokemonToRecord(poke))
		}
		records[name] = entries
	}
	payload := userData{
		User:           c.UserName,
		Pokedex:        records,
		Inventory:      inventoryToRecord(c.Inventory),
		LastDailyGrant: c.LastDailyGrant,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.StoragePath, data, 0o600)
}

func defaultInventory() Inventory {
	return Inventory{
		Pokeball:  10,
		GreatBall: 5,
		UltraBall: 2,
		Potion:    3,
	}
}

func inventoryFromRecord(record inventoryRecord) Inventory {
	return Inventory{
		Pokeball:  record.Pokeball,
		GreatBall: record.GreatBall,
		UltraBall: record.UltraBall,
		Potion:    record.Potion,
	}
}

func inventoryToRecord(inv Inventory) inventoryRecord {
	return inventoryRecord{
		Pokeball:  inv.Pokeball,
		GreatBall: inv.GreatBall,
		UltraBall: inv.UltraBall,
		Potion:    inv.Potion,
	}
}

func pokemonToRecord(pokemon Pokemon) pokemonRecord {
	abilities := make([]pokemonAbilityRecord, 0, len(pokemon.abilities))
	for _, ability := range pokemon.abilities {
		abilities = append(abilities, pokemonAbilityRecord{
			Name:     ability.name,
			IsHidden: ability.isHidden,
			Slot:     ability.slot,
		})
	}
	stats := make(map[string]int, len(pokemon.stats))
	for key, value := range pokemon.stats {
		stats[key] = value
	}
	moves := make([]pokemonMoveRecord, 0, len(pokemon.moves))
	for _, move := range pokemon.moves {
		moves = append(moves, pokemonMoveRecord{
			Name:     move.name,
			Power:    move.power,
			Accuracy: move.accuracy,
			Type:     move.moveType,
			Priority: move.priority,
		})
	}
	return pokemonRecord{
		Name:           pokemon.name,
		DateCaught:     pokemon.dateCaught,
		Height:         pokemon.height,
		Weight:         pokemon.weight,
		Stats:          stats,
		Types:          append([]string(nil), pokemon.types...),
		ID:             pokemon.id,
		BaseExperience: pokemon.baseExperience,
		Order:          pokemon.order,
		IsDefault:      pokemon.isDefault,
		Species:        pokemon.species,
		Abilities:      abilities,
		HeldItems:      append([]string(nil), pokemon.heldItems...),
		Forms:          append([]string(nil), pokemon.forms...),
		Moves:          moves,
		MoveCount:      pokemon.moveCount,
		Level:          pokemon.level,
		Experience:     pokemon.experience,
		GrowthRate:     pokemon.growthRate,
		EvolutionChain: pokemon.evolutionChain,
		LastXPAt:       pokemon.lastXPAt,
		LastXPGain:     pokemon.lastXPGain,
	}
}

func recordToPokemon(record pokemonRecord) Pokemon {
	abilities := make([]pokemonAbility, 0, len(record.Abilities))
	for _, ability := range record.Abilities {
		abilities = append(abilities, pokemonAbility{
			name:     ability.Name,
			isHidden: ability.IsHidden,
			slot:     ability.Slot,
		})
	}
	stats := make(map[string]int, len(record.Stats))
	for key, value := range record.Stats {
		stats[key] = value
	}
	moves := make([]PokemonMove, 0, len(record.Moves))
	for _, move := range record.Moves {
		moves = append(moves, PokemonMove{
			name:     move.Name,
			power:    move.Power,
			accuracy: move.Accuracy,
			moveType: move.Type,
			priority: move.Priority,
		})
	}
	moveCount := record.MoveCount
	if moveCount == 0 {
		moveCount = len(moves)
	}
	level := record.Level
	if level <= 0 {
		level = 1
	}
	return Pokemon{
		name:           record.Name,
		dateCaught:     record.DateCaught,
		height:         record.Height,
		weight:         record.Weight,
		stats:          stats,
		types:          append([]string(nil), record.Types...),
		id:             record.ID,
		baseExperience: record.BaseExperience,
		order:          record.Order,
		isDefault:      record.IsDefault,
		species:        record.Species,
		abilities:      abilities,
		heldItems:      append([]string(nil), record.HeldItems...),
		forms:          append([]string(nil), record.Forms...),
		moves:          moves,
		moveCount:      moveCount,
		level:          level,
		experience:     record.Experience,
		growthRate:     record.GrowthRate,
		evolutionChain: record.EvolutionChain,
		lastXPAt:       record.LastXPAt,
		lastXPGain:     record.LastXPGain,
	}
}

func applyDailyGrant(c *config, now time.Time) (bool, error) {
	if c == nil || c.StoragePath == "" {
		return false, nil
	}
	key := now.Format("2006-01-02")
	if c.LastDailyGrant == key {
		return false, nil
	}
	c.Inventory.Pokeball += 50
	c.Inventory.GreatBall += 20
	c.LastDailyGrant = key
	if err := saveUserData(c); err != nil {
		return false, err
	}
	return true, nil
}
