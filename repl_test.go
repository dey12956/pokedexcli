package main

import (
	"testing"
	"github.com/google/go-cmp/cmp"
)

func TestCleanInput(t *testing.T) {
	tests := map[string]struct {
		input string
		want []string
	}{
		"simple": {input: "hello world", want: []string{"hello", "world"}},
		"simple trailing": {input: "  hello  world  ", want: []string{"hello", "world"}},
		"lowercase": {input: "Charmander Bulbasaur PIKACHU", want: []string{"charmander", "bulbasaur", "pikachu"}},
		"nowhitespace": {input: "helloworld", want: []string{"helloworld"}},
		"nothing": {input: "", want: []string{}},
		"whitespace": {input: "   ", want: []string{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := cleanInput(tc.input)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
