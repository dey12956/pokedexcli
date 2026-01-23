package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func commandMap(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command map doesn't take arguments")
	}

	if c.Next == nil && c.mapFetched {
		return errors.New("You are on the last page")
	}

	locationResp, err := c.pokeapiClient.ListLocations(c.Next)
	if err != nil {
		return err
	}

	c.Next = locationResp.Next
	c.Previous = locationResp.Previous
	c.mapFetched = true

	locations := make([]string, 0, len(locationResp.Results))
	for _, area := range locationResp.Results {
		locations = append(locations, area.Name)
	}

	return promptMapSelection(c, locations)
}

func commandMapB(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command mapb doesn't take arguments")
	}
	if c.Previous == nil {
		return errors.New("You are on the first page")
	}

	locationResp, err := c.pokeapiClient.ListLocations(c.Previous)
	if err != nil {
		return err
	}

	c.Next = locationResp.Next
	c.Previous = locationResp.Previous
	c.mapFetched = true

	locations := make([]string, 0, len(locationResp.Results))
	for _, area := range locationResp.Results {
		locations = append(locations, area.Name)
	}

	return promptMapSelection(c, locations)
}

type mapNavAction int

const (
	mapNavNone mapNavAction = iota
	mapNavNext
	mapNavPrev
)

func promptMapSelection(c *config, locations []string) error {
	if len(locations) == 0 {
		return nil
	}
	for i, name := range locations {
		fmt.Printf("%d) %s\n", i+1, name)
	}
	fmt.Println("Use Left/Right arrows for previous/next page, or enter a number to explore.")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Explore # (or press Enter to continue) > ")
		input, action, needsNewline, err := readMapSelection(reader)
		if err != nil {
			return err
		}
		if needsNewline {
			fmt.Println()
		}
		switch action {
		case mapNavNext:
			return commandMap(c)
		case mapNavPrev:
			return commandMapB(c)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			return nil
		}
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(locations) {
			fmt.Printf("Enter 1-%d, or press Enter to continue.\n", len(locations))
			continue
		}
		return commandExplore(c, locations[choice-1])
	}
}

func readMapSelection(reader *bufio.Reader) (string, mapNavAction, bool, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		input, needsNewline, err := readLine(reader)
		return input, mapNavNone, needsNewline, err
	}
	state, err := term.MakeRaw(fd)
	if err != nil {
		input, needsNewline, err := readLine(reader)
		return input, mapNavNone, needsNewline, err
	}
	defer func() {
		_ = term.Restore(fd, state)
	}()

	var digits strings.Builder
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", mapNavNone, false, err
		}
		switch b {
		case '\r', '\n':
			return digits.String(), mapNavNone, true, nil
		case 27:
			next, err := reader.ReadByte()
			if err != nil {
				return "", mapNavNone, false, err
			}
			if next != '[' {
				continue
			}
			dir, err := reader.ReadByte()
			if err != nil {
				return "", mapNavNone, false, err
			}
			switch dir {
			case 'C':
				return "", mapNavNext, true, nil
			case 'D':
				return "", mapNavPrev, true, nil
			}
		case 8, 127:
			if digits.Len() > 0 {
				value := digits.String()
				digits.Reset()
				digits.WriteString(value[:len(value)-1])
				fmt.Print("\b \b")
			}
		default:
			if b >= '0' && b <= '9' {
				digits.WriteByte(b)
				fmt.Print(string([]byte{b}))
			}
		}
	}
}
