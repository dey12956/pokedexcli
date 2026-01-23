package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type tuiState struct {
	config      *config
	app         *tview.Application
	pages       *tview.Pages
	mapTable    *tview.Table
	pokemonList *tview.List
	asciiView   *tview.TextView
	statusView  *tview.TextView
	pokedexList *tview.List
	pokedexView *tview.TextView
	inspectView *tview.TextView
	locations   []string
	next        *string
	prev        *string
	selected    string
	selectedPkm string
	inspectName string
	spriteCache map[string]string
	focusOnMap  bool
	focusOnDex  bool
	activeView  string
}

const (
	viewMap     = "map"
	viewDex     = "pokedex"
	viewInspect = "inspect"
)

const tuiHelpText = "(h/j/k/l) move  (tab) switch  (c) catch  (m) map  (d) dex  (i) inspect  (n) next  (p) prev  (q) quit"

func commandTui(c *config, name ...string) error {
	if len(name) != 0 {
		return errors.New("Command tui doesn't take arguments")
	}

	app := tview.NewApplication()
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorBlack
	mapTable := tview.NewTable().SetSelectable(true, false)
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
	pokedexList := tview.NewList()
	pokedexList.SetBorder(true).SetTitle("Pokedex")
	pokedexList.SetBackgroundColor(tcell.ColorBlack)
	pokedexView := tview.NewTextView().SetDynamicColors(false)
	pokedexView.SetBorder(true).SetTitle("Details")
	pokedexView.SetBackgroundColor(tcell.ColorBlack)
	inspectView := tview.NewTextView().SetDynamicColors(false)
	inspectView.SetBorder(true).SetTitle("Inspect")
	inspectView.SetBackgroundColor(tcell.ColorBlack)

	state := &tuiState{
		config:      c,
		app:         app,
		pages:       tview.NewPages(),
		mapTable:    mapTable,
		pokemonList: pokemonList,
		asciiView:   asciiView,
		statusView:  statusView,
		pokedexList: pokedexList,
		pokedexView: pokedexView,
		inspectView: inspectView,
		spriteCache: make(map[string]string),
		focusOnMap:  true,
		focusOnDex:  true,
		activeView:  viewMap,
	}

	right := tview.NewFlex().SetDirection(tview.FlexRow)
	right.AddItem(pokemonList, 0, 2, false)
	right.AddItem(asciiView, 0, 3, false)

	main := tview.NewFlex()
	main.AddItem(mapTable, 0, 2, true)
	main.AddItem(right, 0, 3, false)

	pokedexPage := tview.NewFlex()
	pokedexPage.AddItem(pokedexList, 0, 2, true)
	pokedexPage.AddItem(pokedexView, 0, 3, false)

	inspectPage := tview.NewFlex()
	inspectPage.AddItem(inspectView, 0, 1, true)

	state.pages.AddPage(viewMap, main, true, true)
	state.pages.AddPage(viewDex, pokedexPage, true, false)
	state.pages.AddPage(viewInspect, inspectPage, true, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(state.pages, 0, 1, true)
	root.AddItem(statusView, 3, 1, false)

	mapTable.SetSelectionChangedFunc(func(row, column int) {
		state.selectLocation(row, column)
	})

	pokemonList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		state.selectPokemon(mainText)
	})

	pokedexList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		state.selectPokedexPokemon(mainText)
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'h':
				return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'l':
				return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
			}
		}
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
				if state.activeView == viewMap {
					state.loadNextLocations()
				}
				return nil
			case 'p':
				if state.activeView == viewMap {
					state.loadPrevLocations()
				}
				return nil
			case 'c':
				if state.activeView == viewMap {
					state.catchSelectedPokemon()
				}
				return nil
			case 'm':
				state.showMap()
				return nil
			case 'd':
				state.showPokedex()
				return nil
			case 'i':
				state.showInspect()
				return nil
			}
		}
		return event
	})

	state.setStatus("Loading locations...")
	if err := state.loadLocations(nil); err != nil {
		return err
	}
	state.app.SetFocus(mapTable)

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
	switch s.activeView {
	case viewMap:
		if s.focusOnMap {
			s.app.SetFocus(s.pokemonList)
			s.focusOnMap = false
			return
		}
		s.app.SetFocus(s.mapTable)
		s.focusOnMap = true
	case viewDex:
		if s.focusOnDex {
			s.app.SetFocus(s.pokedexView)
			s.focusOnDex = false
			return
		}
		s.app.SetFocus(s.pokedexList)
		s.focusOnDex = true
	}
}

func (s *tuiState) showMap() {
	s.activeView = viewMap
	s.pages.SwitchToPage(viewMap)
	if s.focusOnMap {
		s.app.SetFocus(s.mapTable)
	} else {
		s.app.SetFocus(s.pokemonList)
	}
}

func (s *tuiState) showPokedex() {
	s.activeView = viewDex
	s.pages.SwitchToPage(viewDex)
	s.refreshPokedex()
	if s.focusOnDex {
		s.app.SetFocus(s.pokedexList)
	} else {
		s.app.SetFocus(s.pokedexView)
	}
}

func (s *tuiState) showInspect() {
	s.activeView = viewInspect
	s.pages.SwitchToPage(viewInspect)
	s.renderInspect(s.inspectName)
	s.app.SetFocus(s.inspectView)
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
	cols := 1
	rows := int(math.Ceil(float64(len(s.locations)) / float64(cols)))
	for i, name := range s.locations {
		cell := tview.NewTableCell(name)
		cell.SetExpansion(1)
		s.mapTable.SetCell(i, 0, cell)
	}
	if rows > 0 {
		s.mapTable.Select(0, 0)
	}
}

func (s *tuiState) selectLocation(row, column int) {
	index := row
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
	s.selectedPkm = name
	s.inspectName = name
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

func (s *tuiState) catchSelectedPokemon() {
	if s.selectedPkm == "" {
		s.setStatus("No Pokemon selected")
		return
	}

	s.setStatus(fmt.Sprintf("Throwing a Pokeball at %s...", s.selectedPkm))
	caught, err := attemptCatch(s.config, s.selectedPkm)
	if err != nil {
		s.setStatus(fmt.Sprintf("Error catching Pokemon: %v", err))
		return
	}
	if caught {
		s.setStatus(fmt.Sprintf("%s was caught!", s.selectedPkm))
		s.refreshPokedex()
		return
	}
	s.setStatus(fmt.Sprintf("%s escaped!", s.selectedPkm))
}

func (s *tuiState) refreshPokedex() {
	s.pokedexList.Clear()
	if len(s.config.Pokedex) == 0 {
		s.pokedexView.SetText("No Pokemon caught")
		return
	}

	names := make([]string, 0, len(s.config.Pokedex))
	for name := range s.config.Pokedex {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		s.pokedexList.AddItem(name, "", 0, nil)
	}
	s.pokedexList.SetCurrentItem(0)
}

func (s *tuiState) selectPokedexPokemon(name string) {
	if name == "" {
		s.pokedexView.SetText("No Pokemon selected")
		return
	}
	s.inspectName = name
	s.renderPokedexSummary(name)
}

func (s *tuiState) renderPokedexSummary(name string) {
	poke, exists := s.config.Pokedex[name]
	if !exists {
		s.pokedexView.SetText("Pokemon not found")
		return
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Name: %s\n", poke.name)
	fmt.Fprintf(&b, "Caught: %s\n", poke.dateCaught.Format("2006-01-02 15:04"))
	fmt.Fprintf(&b, "Height: %d\n", poke.height)
	fmt.Fprintf(&b, "Weight: %d\n", poke.weight)
	if len(poke.types) > 0 {
		fmt.Fprintf(&b, "Types: %s\n", strings.Join(poke.types, ", "))
	}

	s.pokedexView.SetText(b.String())
}

func (s *tuiState) renderInspect(name string) {
	if name == "" {
		s.inspectView.SetText("No Pokemon selected")
		return
	}
	poke, exists := s.config.Pokedex[name]
	if !exists {
		s.inspectView.SetText("You have not caught that pokemon")
		return
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Name: %s\n", poke.name)
	fmt.Fprintf(&b, "Height: %d\n", poke.height)
	fmt.Fprintf(&b, "Weight: %d\n", poke.weight)
	fmt.Fprintf(&b, "Caught: %s\n", poke.dateCaught.Format(time.RFC3339))
	fmt.Fprintln(&b, "Stats:")
	fmt.Fprintf(&b, "-hp: %d\n", poke.stats["hp"])
	fmt.Fprintf(&b, "-attack: %d\n", poke.stats["attack"])
	fmt.Fprintf(&b, "-defense: %d\n", poke.stats["defense"])
	fmt.Fprintf(&b, "-special-attack: %d\n", poke.stats["special-attack"])
	fmt.Fprintf(&b, "-special-defense: %d\n", poke.stats["special-defense"])
	fmt.Fprintf(&b, "-speed: %d\n", poke.stats["speed"])
	fmt.Fprintln(&b, "Types:")
	for _, poketype := range poke.types {
		fmt.Fprintf(&b, "-%s\n", poketype)
	}

	s.inspectView.SetText(b.String())
}
