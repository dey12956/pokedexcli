package main

import (
	"errors"
	"fmt"
)

var errExit = errors.New("exit requested")

func commandExit(c *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	return errExit
}
