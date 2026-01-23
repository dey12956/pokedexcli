package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type userData struct {
	User    string                   `json:"user"`
	Pokedex map[string]pokemonRecord `json:"pokedex"`
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
	MoveCount      int                    `json:"move_count"`
}

func promptUserName() (string, error) {
	defaultName := defaultUserName()
	fmt.Print("Trainer name: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
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

func loadUserData(path string) (map[string]Pokemon, error) {
	if path == "" {
		return make(map[string]Pokemon), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]Pokemon), nil
		}
		return nil, err
	}

	var stored userData
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	result := make(map[string]Pokemon, len(stored.Pokedex))
	for name, record := range stored.Pokedex {
		result[name] = recordToPokemon(record)
	}
	return result, nil
}

func saveUserData(c *config) error {
	if c == nil || c.StoragePath == "" {
		return nil
	}
	records := make(map[string]pokemonRecord, len(c.Pokedex))
	for name, poke := range c.Pokedex {
		records[name] = pokemonToRecord(poke)
	}
	payload := userData{
		User:    c.UserName,
		Pokedex: records,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.StoragePath, data, 0o600)
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
		MoveCount:      pokemon.moveCount,
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
		moveCount:      record.MoveCount,
	}
}
