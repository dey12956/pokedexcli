package main

import (
	"errors"
	"fmt"
)

func commandMap(c *config) error {
	if c.Next == nil && c.Previous != nil {
		return errors.New("You are on the last page")
	}

	locationResp, err := c.pokeapiClient.ListLocations(c.Next)
	if err != nil {
		return err
	}

	c.Next = locationResp.Next
	c.Previous = locationResp.Previous

	for _, area := range locationResp.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandMapB(c *config) error {
	if c.Previous == nil {
		return errors.New("You are on the first page")
	}

	locationResp, err := c.pokeapiClient.ListLocations(c.Previous)
	if err != nil {
		return err
	}

	c.Next = locationResp.Next
	c.Previous = locationResp.Previous

	for _, area := range locationResp.Results {
		fmt.Println(area.Name)
	}

	return nil
}
