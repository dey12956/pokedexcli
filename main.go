package main

import (
	"github.com/dey12956/pokedexcli/internal/pokeapi"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	pokeClient := pokeapi.NewClient(5*time.Second, 5*time.Minute)
	c := &config{
		pokeapiClient: pokeClient,
		Pokedex:       make(map[string]Pokemon),
	}
	startRepl(c)
}
