package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

func main() {
	rand.Seed(time.Now().UnixNano())
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
		Pokedex:       make(map[string]Pokemon),
		UserName:      userName,
		StoragePath:   storagePath,
	}
	if storagePath != "" {
		loaded, err := loadUserData(storagePath)
		if err != nil {
			fmt.Printf("Warning: failed to load data: %v\n", err)
		} else {
			c.Pokedex = loaded
		}
	}

	fmt.Printf("Using trainer: %s\n", userName)
	startRepl(c)
}
