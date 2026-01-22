package main

import(
	"strings"
)

func cleanInput(text string) []string {
	lowerCaseString := strings.ToLower(text)
	return strings.Fields(lowerCaseString)
}
