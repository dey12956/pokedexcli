package main

import(
	"bufio"
	"fmt"
	"os"
	"strings"
)

func startRepl() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() { break }

		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 0 { continue }

		if command, exists := getCommands()[words[0]]; exists{
			err := command.callback()
			if err != nil {
				fmt.Println(err)
			}
			continue
		} else {
			fmt.Println("Unknown command")
			continue
		}
	}
}

func cleanInput(text string) []string {
	lowerCaseString := strings.ToLower(text)
	return strings.Fields(lowerCaseString)
}

type cliCommand struct {
	name string
	description string
	callback func() error
}

func getCommands() map[string]cliCommand{
	return map[string]cliCommand{
		"exit": {
			name: "exit",
			description: "Exit the Pokedex",
			callback: commandExit,
		},
		"help": {
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
	}
}

