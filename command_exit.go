package main

import (
	"errors"
	"fmt"
)

var errExit = errors.New("exit requested")

func commandExit(c *config, name ...string) error {
	fmt.Println()
	fmt.Println("Closing the Pokedex... Goodbye!")
	fmt.Println()
	return errExit
}
