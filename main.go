package main

import (
	"time"
	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

func main() {
	pokeClient := pokeapi.NewClient(5 * time.Second)
	c := &config {
		pokeapiClient: pokeClient,
	}
	startRepl(c)
}
