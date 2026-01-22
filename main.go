package main

import (
	"bufio"
	"os"
	"fmt"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {break}
		input := scanner.Text()
		words := cleanInput(input)
		if len(words) == 0 { continue }
		fmt.Printf("Your command was: %s\n", words[0])
	}
}
