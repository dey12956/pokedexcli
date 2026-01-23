package main

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type tuiState struct {
	config      *config
	app         *tview.Application
	mapTable    *tview.Table
	pokemonList *tview.List
	asciiView   *tview.TextView
	statusView  *tview.TextView
	locations   []string
	next        *string
	prev        *string
	selected    string
	spriteCache map[string]string
	focusOnMap  bool
}

const tuiHelpText = "(n) next  (p) prev  (tab) switch  (q) quit"

func commandTui(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command tui doesn't take arguments")
	}

	app := tview.NewApplication()
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorBlack
	mapTable := tview.NewTable().SetSelectable(true, true)
	mapTable.SetBorder(true).SetTitle("Map")
	mapTable.SetBackgroundColor(tcell.ColorBlack)
	pokemonList := tview.NewList()
	pokemonList.SetBorder(true).SetTitle("Pokemon")
	pokemonList.SetBackgroundColor(tcell.ColorBlack)
	asciiView := tview.NewTextView().SetDynamicColors(false)
	asciiView.SetBorder(true).SetTitle("Sprite")
	asciiView.SetBackgroundColor(tcell.ColorBlack)
	statusView := tview.NewTextView().SetDynamicColors(true)
	statusView.SetBorder(true).SetTitle("Status")
	statusView.SetBackgroundColor(tcell.ColorBlack)

	state := &tuiState{
		config:      c,
		app:         app,
		mapTable:    mapTable,
		pokemonList: pokemonList,
		asciiView:   asciiView,
		statusView:  statusView,
		spriteCache: make(map[string]string),
		focusOnMap:  true,
	}

	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(pokemonList, 0, 2, false)
	right.AddItem(asciiView, 0, 3, false)

	main := tview.NewFlex()
	main.AddItem(mapTable, 0, 2, true)
	main.AddItem(right, 0, 3, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(main, 0, 1, true)
	root.AddItem(statusView, 3, 1, false)

	mapTable.SetSelectionChangedFunc(func(row, column int) {
		state.selectLocation(row, column)
	})

	pokemonList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		state.selectPokemon(mainText)
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			app.Stop()
			return nil
		case tcell.KeyTab:
			state.toggleFocus()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'n':
				state.loadNextLocations()
				return nil
			case 'p':
				state.loadPrevLocations()
				return nil
			}
		}
		return event
	})

	state.setStatus("Loading locations...")
	if err := state.loadLocations(nil); err != nil {
		return err
	}

	if err := app.SetRoot(root, true).EnableMouse(false).Run(); err != nil {
		return err
	}

	return nil
}

func (s *tuiState) setStatus(text string) {
	if text == "" {
		s.statusView.SetText(tuiHelpText)
		return
	}
	s.statusView.SetText(fmt.Sprintf("%s  %s", text, tuiHelpText))
}

func (s *tuiState) toggleFocus() {
	if s.focusOnMap {
		s.app.SetFocus(s.pokemonList)
		s.focusOnMap = false
		return
	}
	s.app.SetFocus(s.mapTable)
	s.focusOnMap = true
}

func (s *tuiState) loadNextLocations() {
	if s.next == nil && s.locations != nil {
		s.setStatus("You are on the last page")
		return
	}
	_ = s.loadLocations(s.next)
}

func (s *tuiState) loadPrevLocations() {
	if s.prev == nil && s.locations != nil {
		s.setStatus("You are on the first page")
		return
	}
	_ = s.loadLocations(s.prev)
}

func (s *tuiState) loadLocations(pageURL *string) error {
	resp, err := s.config.pokeapiClient.ListLocations(pageURL)
	if err != nil {
		s.setStatus(fmt.Sprintf("Error loading locations: %v", err))
		return err
	}

	s.next = resp.Next
	s.prev = resp.Previous
	s.locations = make([]string, 0, len(resp.Results))
	for _, area := range resp.Results {
		s.locations = append(s.locations, area.Name)
	}

	s.selected = ""
	s.renderLocations()
	if len(s.locations) == 0 {
		s.pokemonList.Clear()
		s.asciiView.SetText("No locations")
		return nil
	}

	s.setStatus("Locations loaded")
	return nil
}

func (s *tuiState) renderLocations() {
	s.mapTable.Clear()
	cols := 3
	rows := int(math.Ceil(float64(len(s.locations)) / float64(cols)))
	for i, name := range s.locations {
		row := i / cols
		col := i % cols
		cell := tview.NewTableCell(strings.ReplaceAll(name, "-", " "))
		s.mapTable.SetCell(row, col, cell)
	}
	if rows > 0 {
		s.mapTable.Select(0, 0)
	}
}

func (s *tuiState) selectLocation(row, column int) {
	cols := 3
	index := row*cols + column
	if index < 0 || index >= len(s.locations) {
		return
	}
	area := s.locations[index]
	if area == s.selected {
		return
	}
	s.selected = area
	s.loadPokemon(area)
}

func (s *tuiState) loadPokemon(area string) {
	s.setStatus(fmt.Sprintf("Loading Pokemon in %s...", area))
	resp, err := s.config.pokeapiClient.ListPokemon(area)
	if err != nil {
		s.setStatus(fmt.Sprintf("Error loading Pokemon: %v", err))
		return
	}

	s.pokemonList.Clear()
	if len(resp.PokemonEncounters) == 0 {
		s.asciiView.SetText("No Pokemon found")
		return
	}

	for _, encounter := range resp.PokemonEncounters {
		name := encounter.Pokemon.Name
		s.pokemonList.AddItem(name, "", 0, nil)
	}
	s.pokemonList.SetCurrentItem(0)
}

func (s *tuiState) selectPokemon(name string) {
	if name == "" {
		return
	}
	if cached, exists := s.spriteCache[name]; exists {
		s.asciiView.SetText(cached)
		return
	}

	s.setStatus(fmt.Sprintf("Loading sprite for %s...", name))
	poke, err := s.config.pokeapiClient.GetPokemon(name)
	if err != nil {
		s.setStatus(fmt.Sprintf("Error loading sprite: %v", err))
		return
	}

	url := spriteURLFromPokemon(poke)
	art, err := fetchSpriteASCII(url)
	if err != nil {
		s.setStatus(fmt.Sprintf("Error converting sprite: %v", err))
		return
	}

	s.spriteCache[name] = art
	s.asciiView.SetText(art)
	s.setStatus("Sprite loaded")
}
