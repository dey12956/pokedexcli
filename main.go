package main

import (
	"github.com/dey12956/pokedexcli/internal/pokeapi"
	"time"
)

func main() {
	pokeClient := pokeapi.NewClient(5*time.Second, 5*time.Minute)
	c := &config{
		pokeapiClient: pokeClient,
		Pokedex:       make(map[string]Pokemon),
	}
	startRepl(c)
}
