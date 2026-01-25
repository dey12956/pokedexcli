package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

func main() {
	pokeClient := pokeapi.NewClient(5*time.Second, 5*time.Minute)

	userName, err := promptUserName()
	if err != nil {
		fmt.Printf("Warning: failed to read trainer name: %v\n", err)
		userName = defaultUserName()
	}
	storagePath, err := userDataPath(userName)
	if err != nil {
		fmt.Printf("Warning: failed to set storage path: %v\n", err)
	}

	c := &config{
		pokeapiClient: pokeClient,
		Pokedex:       make(map[string][]Pokemon),
		Inventory:     defaultInventory(),
		UserName:      userName,
		StoragePath:   storagePath,
	}
	if storagePath != "" {
		dataExists := false
		if _, err := os.Stat(storagePath); err == nil {
			dataExists = true
		}
		loaded, inventory, lastDaily, err := loadUserData(storagePath)
		if err != nil {
			fmt.Printf("Warning: failed to load data: %v\n", err)
		} else {
			c.Pokedex = loaded
			if inventory != (Inventory{}) {
				c.Inventory = inventory
			}
			c.LastDailyGrant = lastDaily
		}
		if err := ensureStarterPokemon(c, dataExists); err != nil {
			fmt.Printf("Warning: failed to add starter: %v\n", err)
		}
	}

	fmt.Printf("Using trainer: %s\n", userName)
	if granted, err := applyDailyGrant(c, time.Now()); err != nil {
		fmt.Printf("Warning: failed to apply daily grant: %v\n", err)
	} else if granted {
		fmt.Println("Daily supply: +50 Pokeballs, +20 Great Balls")
	}
	startRepl(c)
}
