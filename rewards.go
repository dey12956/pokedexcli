package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func grantRandomSupplies(c *config, reason string) {
	if c == nil {
		return
	}
	red, blue, black := randomBallReward()
	c.Inventory.Pokeball += red
	c.Inventory.GreatBall += blue
	c.Inventory.UltraBall += black
	c.Inventory.Potion += 3
	label := strings.TrimSpace(reason)
	if label == "" {
		label = "reward"
	}
	fmt.Printf("%s reward: +%d Pokeballs, +%d Great Balls, +%d Ultra Balls, +3 Potions.\n", label, red, blue, black)
	saveUserData(c)
}

func randomBallReward() (int, int, int) {
	const total = 10
	const types = 3
	balls := [types]int{1, 1, 1}
	for i := 0; i < total-types; i++ {
		balls[rand.Intn(types)]++
	}
	return balls[0], balls[1], balls[2]
}
