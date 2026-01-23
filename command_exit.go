package main

import (
	"errors"
	"fmt"
)

var errExit = errors.New("exit requested")

func commandExit(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command exit doesn't take arguments")
	}
	fmt.Println()
	fmt.Println("Closing the Pokedex... Goodbye!")
	fmt.Println()
	return errExit
}
