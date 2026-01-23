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
	config        *config
	app           *tview.Application
	pages         *tview.Pages
	mapTable      *tview.Table
	pokemonList   *tview.List
	asciiView     *tview.TextView
	statusView    *tview.TextView
	root          *tview.Flex
	pokedexList   *tview.List
	pokedexView   *tview.TextView
	inspectView   *tview.TextView
	locations     []string
	next          *string
	prev          *string
	selected      string
	selectedPkm   string
	inspectName   string
	spriteCache   map[string]string
	focusOnMap    bool
	focusOnDex    bool
	activeView    string
	statusMessage string
}

const (
	viewMap     = "map"
	viewDex     = "pokedex"
	viewInspect = "inspect"
)

const tuiHelpTextFull = "(h/j/k/l) move (tab) switch (c) catch (m) map (d) dex (i) inspect (n) next (p) prev (q) quit"
const tuiHelpTextShort = "Window small: resize for full help. (h/j/k/l) move (tab) switch (c) catch (m) map (d) dex (i) inspect (q) quit"

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
	statusView.SetWrap(true)
	statusView.SetWordWrap(true)
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
	state.root = root

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

	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		state.updateStatusView(width)
		return false
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
	s.statusMessage = text
	s.updateStatusView(0)
}

func (s *tuiState) updateStatusView(width int) {
	if s.statusView == nil || s.root == nil {
		return
	}
	if width <= 0 {
		_, _, currentWidth, _ := s.statusView.GetRect()
		width = currentWidth
	}
	if width <= 0 {
		width = 80
	}
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	helpText := tuiHelpTextFull
	if wrappedLineCount(tuiHelpTextFull, innerWidth) > 2 {
		helpText = tuiHelpTextShort
	}

	text := helpText
	if strings.TrimSpace(s.statusMessage) != "" {
		text = fmt.Sprintf("%s %s", s.statusMessage, helpText)
	}

	s.statusView.SetText(text)
	height := wrappedLineCount(text, innerWidth) + 2
	if height < 3 {
		height = 3
	}
	s.root.ResizeItem(s.statusView, height, 0)
}

func wrappedLineCount(text string, width int) int {
	if width <= 0 {
		return 1
	}
	lines := 0
	for _, rawLine := range strings.Split(text, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			lines++
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			lines++
			continue
		}
		lineCount := 1
		lineLen := 0
		for _, word := range words {
			wordLen := len(word)
			if wordLen > width {
				if lineLen > 0 {
					lineCount++
					lineLen = 0
				}
				for wordLen > width {
					lineCount++
					wordLen -= width
				}
				if wordLen == 0 {
					lineLen = 0
					continue
				}
				lineLen = wordLen
				continue
			}
			if lineLen == 0 {
				lineLen = wordLen
				continue
			}
			if lineLen+1+wordLen <= width {
				lineLen += 1 + wordLen
				continue
			}
			lineCount++
			lineLen = wordLen
		}
		lines += lineCount
	}
	if lines == 0 {
		return 1
	}
	return lines
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
		if err := saveUserData(s.config); err != nil {
			s.setStatus(fmt.Sprintf("%s was caught, but failed to save: %v", s.selectedPkm, err))
		} else {
			s.setStatus(fmt.Sprintf("%s was caught!", s.selectedPkm))
		}
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
	fmt.Fprintf(&b, "ID: %d\n", poke.id)
	if poke.species != "" {
		fmt.Fprintf(&b, "Species: %s\n", poke.species)
	}
	fmt.Fprintf(&b, "Base XP: %d\n", poke.baseExperience)
	fmt.Fprintf(&b, "Height: %d\n", poke.height)
	fmt.Fprintf(&b, "Weight: %d\n", poke.weight)
	fmt.Fprintf(&b, "Moves: %d\n", poke.moveCount)
	if len(poke.types) > 0 {
		fmt.Fprintf(&b, "Types: %s\n", strings.Join(poke.types, ", "))
	}
	if len(poke.abilities) > 0 {
		abilities := make([]string, 0, len(poke.abilities))
		for _, ability := range poke.abilities {
			label := ability.name
			if ability.isHidden {
				label = fmt.Sprintf("%s (hidden)", label)
			}
			abilities = append(abilities, label)
		}
		fmt.Fprintf(&b, "Abilities: %s\n", strings.Join(abilities, ", "))
	}
	if len(poke.heldItems) > 0 {
		fmt.Fprintf(&b, "Held items: %s\n", strings.Join(poke.heldItems, ", "))
	} else {
		fmt.Fprintln(&b, "Held items: none")
	}
	if len(poke.forms) > 0 {
		fmt.Fprintf(&b, "Forms: %s\n", strings.Join(poke.forms, ", "))
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
	fmt.Fprintf(&b, "ID: %d\n", poke.id)
	if poke.species != "" {
		fmt.Fprintf(&b, "Species: %s\n", poke.species)
	}
	fmt.Fprintf(&b, "Base XP: %d\n", poke.baseExperience)
	fmt.Fprintf(&b, "Height: %d\n", poke.height)
	fmt.Fprintf(&b, "Weight: %d\n", poke.weight)
	fmt.Fprintf(&b, "Order: %d\n", poke.order)
	fmt.Fprintf(&b, "Default: %t\n", poke.isDefault)
	fmt.Fprintf(&b, "Caught: %s\n", poke.dateCaught.Format(time.RFC3339))
	fmt.Fprintf(&b, "Moves: %d\n", poke.moveCount)
	fmt.Fprintln(&b, "Abilities:")
	if len(poke.abilities) == 0 {
		fmt.Fprintln(&b, "-none")
	} else {
		for _, ability := range poke.abilities {
			label := fmt.Sprintf("-%s (slot %d)", ability.name, ability.slot)
			if ability.isHidden {
				label = fmt.Sprintf("%s hidden", label)
			}
			fmt.Fprintln(&b, label)
		}
	}
	fmt.Fprintln(&b, "Held items:")
	if len(poke.heldItems) == 0 {
		fmt.Fprintln(&b, "-none")
	} else {
		for _, item := range poke.heldItems {
			fmt.Fprintf(&b, "-%s\n", item)
		}
	}
	if len(poke.forms) > 0 {
		fmt.Fprintln(&b, "Forms:")
		for _, form := range poke.forms {
			fmt.Fprintf(&b, "-%s\n", form)
		}
	}
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
