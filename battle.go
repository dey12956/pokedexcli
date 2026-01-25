package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

type battlePokemon struct {
	pokemon Pokemon
	current int
	max     int
}

type pokemonSelection struct {
	key   string
	index int
	label string
}

var errSelectionCancelled = errors.New("selection cancelled")

const (
	statusNone      = "none"
	statusSleep     = "sleep"
	statusParalysis = "paralysis"
)

func commandBattle(c *config, name ...string) error {
	if len(name) == 0 {
		return errors.New("Enter a Pokemon to battle")
	}
	if len(name) > 1 {
		return errors.New("Command battle takes a single Pokemon")
	}
	if len(c.Pokedex) == 0 {
		return errors.New("No Pokemon in your Pokedex")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("A wild %s appeared!\n", name[0])

	wildResp, err := c.pokeapiClient.GetPokemon(name[0])
	if err != nil {
		return err
	}
	wild, err := buildPokemonFromResponse(c, wildResp)
	if err != nil {
		return err
	}
	wild.dateCaught = time.Time{}
	wildBattle := battlePokemon{pokemon: wild, max: maxHP(wild)}
	wildBattle.current = wildBattle.max
	wildStatus := statusNone
	var selection pokemonSelection
	var playerBattle battlePokemon
	playerSelected := false

	round := 1
	for {
		fmt.Printf("\nRound %d\n", round)
		if playerSelected {
			fmt.Printf("Your %s HP: %d/%d\n", playerBattle.pokemon.name, playerBattle.current, playerBattle.max)
		} else {
			fmt.Println("Your Pokemon: (not selected)")
		}
		fmt.Printf("Wild %s HP: %d/%d\n", wildBattle.pokemon.name, wildBattle.current, wildBattle.max)

		action, cancelled, err := promptChoice(reader, "Choose action: 1) Fight 2) Catch 3) Run 4) Item > ", 4)
		if err != nil {
			return err
		}
		if cancelled {
			fmt.Println("Battle cancelled")
			return nil
		}

		switch action {
		case 1:
			if !playerSelected {
				selection, err = choosePlayerPokemon(c, reader)
				if err != nil {
					if errors.Is(err, errSelectionCancelled) {
						return nil
					}
					return err
				}
				player := c.Pokedex[selection.key][selection.index]
				if err := applyRestXP(c, &player); err != nil {
					return err
				}
				playerBattle = battlePokemon{pokemon: player, max: maxHP(player)}
				playerBattle.current = playerBattle.max
				playerSelected = true
			}
			move := chooseMove(reader, playerBattle.pokemon)
			wildMove := chooseWildMove(wildBattle.pokemon)
			playerFirst := decideFirst(move, wildMove, playerBattle.pokemon, wildBattle.pokemon)
			if playerFirst {
				resolveAttack(&playerBattle, &wildBattle, move, true)
				if wildBattle.current <= 0 {
					fmt.Printf("Wild %s fainted!\n", wildBattle.pokemon.name)
					err = awardBattleXP(c, &playerBattle.pokemon, wildBattle.pokemon.baseExperience)
					if err == nil {
						syncPlayerPokemon(c, selection, playerBattle.pokemon)
						saveUserData(c)
					}
					grantRandomSupplies(c, "Battle win")
					return err
				}
				resolveAttack(&wildBattle, &playerBattle, wildMove, false)
			} else {
				resolveAttack(&wildBattle, &playerBattle, wildMove, false)
				if playerBattle.current <= 0 {
					fmt.Printf("%s fainted!\n", playerBattle.pokemon.name)
					syncPlayerPokemon(c, selection, playerBattle.pokemon)
					saveUserData(c)
					return nil
				}
				resolveAttack(&playerBattle, &wildBattle, move, true)
			}
			if playerBattle.current <= 0 {
				fmt.Printf("%s fainted!\n", playerBattle.pokemon.name)
				syncPlayerPokemon(c, selection, playerBattle.pokemon)
				saveUserData(c)
				return nil
			}
			if wildBattle.current <= 0 {
				fmt.Printf("Wild %s fainted!\n", wildBattle.pokemon.name)
				err = awardBattleXP(c, &playerBattle.pokemon, wildBattle.pokemon.baseExperience)
				if err == nil {
					syncPlayerPokemon(c, selection, playerBattle.pokemon)
					saveUserData(c)
				}
				grantRandomSupplies(c, "Battle win")
				return err
			}
		case 2:
			caught, err := attemptCatchInBattle(reader, c, &wildBattle, wildStatus)
			if err != nil {
				return err
			}
			if caught {
				fmt.Printf("%s was caught!\n", wildBattle.pokemon.name)
				appendCaughtPokemon(c, wildBattle.pokemon)
				if playerSelected {
					err = awardCaptureXP(c, &playerBattle.pokemon, wildBattle.pokemon.baseExperience)
					if err == nil {
						syncPlayerPokemon(c, selection, playerBattle.pokemon)
						saveUserData(c)
					}
				} else {
					saveUserData(c)
				}
				grantRandomSupplies(c, "Catch")
				return err
			}
			fmt.Printf("%s escaped the ball!\n", wildBattle.pokemon.name)
			if playerSelected {
				wildMove := chooseWildMove(wildBattle.pokemon)
				resolveAttack(&wildBattle, &playerBattle, wildMove, false)
				if playerBattle.current <= 0 {
					fmt.Printf("%s fainted!\n", playerBattle.pokemon.name)
					syncPlayerPokemon(c, selection, playerBattle.pokemon)
					saveUserData(c)
					return nil
				}
			}
		case 3:
			fmt.Println("You ran away.")
			if playerSelected {
				syncPlayerPokemon(c, selection, playerBattle.pokemon)
				saveUserData(c)
			}
			return nil
		case 4:
			status, err := applyPotion(reader, c)
			if err != nil {
				return err
			}
			if status != "" {
				wildStatus = status
				fmt.Printf("Wild %s is now %s.\n", wildBattle.pokemon.name, wildStatus)
			}
		}

		round++
	}
}

func choosePlayerPokemon(c *config, reader *bufio.Reader) (pokemonSelection, error) {
	if len(c.Pokedex) == 0 {
		return pokemonSelection{}, errNoPokemon
	}

	keys := make([]string, 0, len(c.Pokedex))
	for key := range c.Pokedex {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	selections := make([]pokemonSelection, 0)
	for _, key := range keys {
		entries := c.Pokedex[key]
		for index, entry := range entries {
			label := fmt.Sprintf("%s (Lv %d)", entry.name, entry.level)
			selections = append(selections, pokemonSelection{key: key, index: index, label: label})
		}
	}

	if len(selections) == 0 {
		return pokemonSelection{}, errNoPokemon
	}

	fmt.Println("Choose your Pokemon:")
	for i, selection := range selections {
		fmt.Printf("%d) %s\n", i+1, selection.label)
	}

	choice, cancelled, err := promptChoice(reader, "Selection > ", len(selections))
	if err != nil {
		return pokemonSelection{}, err
	}
	if cancelled {
		return pokemonSelection{}, errSelectionCancelled
	}
	return selections[choice-1], nil
}

func chooseMove(reader *bufio.Reader, pokemon Pokemon) PokemonMove {
	moves := availableMoves(pokemon)
	fmt.Println("Choose a move:")
	for i, move := range moves {
		fmt.Printf("%d) %s (power %d, acc %d, prio %d, type %s)\n", i+1, move.name, move.power, move.accuracy, move.priority, move.moveType)
	}
	choice, cancelled, err := promptChoice(reader, "Move > ", len(moves))
	if err != nil || cancelled {
		return moves[0]
	}
	return moves[choice-1]
}

func chooseWildMove(pokemon Pokemon) PokemonMove {
	moves := availableMoves(pokemon)
	return moves[rng.Intn(len(moves))]
}

func availableMoves(pokemon Pokemon) []PokemonMove {
	if len(pokemon.moves) == 0 {
		return []PokemonMove{{name: "tackle", power: 40, accuracy: 100, priority: 0, moveType: "normal"}}
	}
	if len(pokemon.moves) > 4 {
		return pokemon.moves[:4]
	}
	return pokemon.moves
}

func decideFirst(playerMove, wildMove PokemonMove, player Pokemon, wild Pokemon) bool {
	if playerMove.priority > wildMove.priority {
		return true
	}
	if playerMove.priority < wildMove.priority {
		return false
	}
	playerSpeed := player.stats["speed"]
	wildSpeed := wild.stats["speed"]
	if playerSpeed == wildSpeed {
		return rng.Intn(2) == 0
	}
	return playerSpeed > wildSpeed
}

func resolveAttack(attacker, defender *battlePokemon, move PokemonMove, isPlayer bool) {
	accuracy := move.accuracy
	if accuracy <= 0 {
		accuracy = 100
	}
	if rng.Intn(100) >= accuracy {
		if isPlayer {
			fmt.Printf("%s used %s but missed!\n", attacker.pokemon.name, move.name)
		} else {
			fmt.Printf("Wild %s used %s but missed!\n", attacker.pokemon.name, move.name)
		}
		return
	}

	damage := calculateDamage(attacker.pokemon, defender.pokemon, move)
	defender.current -= damage
	if defender.current < 0 {
		defender.current = 0
	}
	if isPlayer {
		fmt.Printf("%s used %s for %d damage!\n", attacker.pokemon.name, move.name, damage)
	} else {
		fmt.Printf("Wild %s used %s for %d damage!\n", attacker.pokemon.name, move.name, damage)
	}
}

func calculateDamage(attacker, defender Pokemon, move PokemonMove) int {
	power := move.power
	if power <= 0 {
		power = 40
	}
	level := attacker.level
	if level <= 0 {
		level = 5
	}
	attack := attacker.stats["attack"]
	defense := defender.stats["defense"]
	base := (power / 3) + (level / 2)
	bonus := (attack / 8) - (defense / 16)
	damage := base + bonus
	return max(1, damage)
}

func maxHP(pokemon Pokemon) int {
	hp := pokemon.stats["hp"]
	if hp <= 0 {
		hp = 50
	}
	level := pokemon.level
	if level <= 0 {
		level = 5
	}
	return hp + (level * 2)
}

func attemptCatchInBattle(reader *bufio.Reader, c *config, wild *battlePokemon, status string) (bool, error) {
	ballChoice, cancelled, err := promptChoice(
		reader,
		fmt.Sprintf(
			"Choose ball: 1) Pokeball (x%d) 2) Great Ball (x%d) 3) Ultra Ball (x%d) > ",
			c.Inventory.Pokeball,
			c.Inventory.GreatBall,
			c.Inventory.UltraBall,
		),
		3,
	)
	if err != nil {
		return false, err
	}
	if cancelled {
		return false, nil
	}

	ballFactor := 1.0
	switch ballChoice {
	case 1:
		if c.Inventory.Pokeball <= 0 {
			fmt.Println("No Pokeballs left")
			return false, nil
		}
		c.Inventory.Pokeball--
		saveUserData(c)
		ballFactor = 0.7
	case 2:
		if c.Inventory.GreatBall <= 0 {
			fmt.Println("No Great Balls left")
			return false, nil
		}
		c.Inventory.GreatBall--
		saveUserData(c)
		ballFactor = 1.0
	case 3:
		if c.Inventory.UltraBall <= 0 {
			fmt.Println("No Ultra Balls left")
			return false, nil
		}
		c.Inventory.UltraBall--
		saveUserData(c)
		ballFactor = 1.15
	}

	statusFactor := 1.0
	if status == statusSleep || status == statusParalysis {
		statusFactor = 1.25
	}
	if wild.max <= 0 {
		wild.max = 1
	}
	hpRatio := float64(wild.current) / float64(wild.max)
	hpFactor := 0.3 + (0.7 * (1.0 - hpRatio))
	base := catchProb(wild.pokemon.baseExperience)
	chance := base * ballFactor * statusFactor * hpFactor
	chance = math.Min(0.95, math.Max(0.02, chance))

	return rng.Float64() < chance, nil
}

func applyPotion(reader *bufio.Reader, c *config) (string, error) {
	if c.Inventory.Potion <= 0 {
		fmt.Println("No potions left")
		return "", nil
	}
	choice, cancelled, err := promptChoice(
		reader,
		fmt.Sprintf("Choose status (Potion x%d): 1) Sleep 2) Paralysis > ", c.Inventory.Potion),
		2,
	)
	if err != nil {
		return "", err
	}
	if cancelled {
		return "", nil
	}
	c.Inventory.Potion--
	saveUserData(c)
	if choice == 1 {
		return statusSleep, nil
	}
	return statusParalysis, nil
}

func awardBattleXP(c *config, pokemon *Pokemon, baseXP int) error {
	if baseXP <= 0 {
		return nil
	}
	gained := int(math.Round(float64(baseXP) * 0.9))
	return applyExperience(c, pokemon, max(1, gained), true)
}

func awardCaptureXP(c *config, pokemon *Pokemon, baseXP int) error {
	if baseXP <= 0 {
		return nil
	}
	return applyExperience(c, pokemon, baseXP, true)
}

func syncPlayerPokemon(c *config, selection pokemonSelection, updated Pokemon) {
	if c == nil {
		return
	}
	entries, exists := c.Pokedex[selection.key]
	if !exists || selection.index >= len(entries) {
		return
	}
	if updated.name == selection.key {
		entries[selection.index] = updated
		c.Pokedex[selection.key] = entries
		return
	}

	entries = append(entries[:selection.index], entries[selection.index+1:]...)
	if len(entries) == 0 {
		delete(c.Pokedex, selection.key)
	} else {
		c.Pokedex[selection.key] = entries
	}
	appendCaughtPokemon(c, updated)
}

func appendCaughtPokemon(c *config, pokemon Pokemon) {
	if c.Pokedex == nil {
		c.Pokedex = make(map[string][]Pokemon)
	}
	key := pokemon.name
	c.Pokedex[key] = append(c.Pokedex[key], pokemon)
}

func promptChoice(reader *bufio.Reader, prompt string, max int) (int, bool, error) {
	for {
		fmt.Printf("\n%s", prompt)
		input, needsNewline, err := readChoice(reader, max)
		if err != nil {
			return 0, false, err
		}
		if needsNewline {
			fmt.Println()
		}
		input = strings.TrimSpace(strings.ToLower(input))
		switch input {
		case "c":
			return 0, true, nil
		case "y":
			if max >= 1 {
				return 1, false, nil
			}
		case "n":
			if max >= 2 {
				return 2, false, nil
			}
		}
		if choice, err := strconv.Atoi(input); err == nil {
			if choice >= 1 && choice <= max {
				return choice, false, nil
			}
		}
		fmt.Printf("Enter 1-%d, y/n, or c to cancel.\n", max)
	}
}

func readChoice(reader *bufio.Reader, max int) (string, bool, error) {
	if max <= 9 {
		fd := int(os.Stdin.Fd())
		if term.IsTerminal(fd) {
			state, err := term.MakeRaw(fd)
			if err == nil {
				defer func() {
					_ = term.Restore(fd, state)
				}()
				b, err := reader.ReadByte()
				if err != nil {
					return "", false, err
				}
				switch b {
				case 'y', 'Y', 'n', 'N', 'c', 'C':
					fmt.Print(string([]byte{b}))
				}
				return string([]byte{b}), true, nil
			}
		}
	}
	return readLine(reader)
}

func readLine(reader *bufio.Reader) (string, bool, error) {
	var builder strings.Builder
	var sawCR bool
	var sawCRLF bool
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) && builder.Len() > 0 {
				return builder.String(), false, nil
			}
			return "", false, err
		}
		if b == '\n' {
			break
		}
		if b == '\r' {
			sawCR = true
			if reader.Buffered() > 0 {
				next, err := reader.Peek(1)
				if err == nil && len(next) == 1 && next[0] == '\n' {
					_, _ = reader.ReadByte()
					sawCRLF = true
				}
			}
			break
		}
		builder.WriteByte(b)
	}
	return builder.String(), sawCR && !sawCRLF, nil
}
