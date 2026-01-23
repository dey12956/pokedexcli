package main

import (
	"time"
	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

func main() {
	pokeClient := pokeapi.NewClient(5 * time.Second, 5 * time.Minute)
	c := &config {
		pokeapiClient: pokeClient,
	}
	startRepl(c)
}
