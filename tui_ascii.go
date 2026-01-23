package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"time"

	"github.com/dey12956/pokedexcli/internal/pokeapi"
)

const asciiWidth = 40

func spriteURLFromPokemon(poke pokeapi.CatchPokemonResponse) string {
	gen1 := poke.Sprites.Versions.GenerationI
	if gen1.RedBlue.FrontDefault != "" {
		return gen1.RedBlue.FrontDefault
	}
	if gen1.Yellow.FrontDefault != "" {
		return gen1.Yellow.FrontDefault
	}

	gen2 := poke.Sprites.Versions.GenerationIi
	if gen2.Crystal.FrontDefault != "" {
		return gen2.Crystal.FrontDefault
	}
	if gen2.Gold.FrontDefault != "" {
		return gen2.Gold.FrontDefault
	}
	if gen2.Silver.FrontDefault != "" {
		return gen2.Silver.FrontDefault
	}

	if poke.Sprites.FrontDefault != "" {
		return poke.Sprites.FrontDefault
	}
	if poke.Sprites.Other.Showdown.FrontDefault != "" {
		return poke.Sprites.Other.Showdown.FrontDefault
	}

	return ""
}

func fetchSpriteASCII(spriteURL string) (string, error) {
	if spriteURL == "" {
		return "", errors.New("no sprite URL available")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(spriteURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("sprite fetch error: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return imageToASCII(img, asciiWidth), nil
}

func imageToASCII(img image.Image, targetWidth int) string {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return ""
	}
	if targetWidth <= 0 {
		targetWidth = width
	}

	aspect := float64(height) / float64(width)
	targetHeight := int(float64(targetWidth) * aspect * 0.5)
	if targetHeight < 1 {
		targetHeight = 1
	}

	chars := []byte(" .:-=+*#%@")
	var buf bytes.Buffer

	for y := 0; y < targetHeight; y++ {
		sy := bounds.Min.Y + int(float64(y)/float64(targetHeight)*float64(height))
		for x := 0; x < targetWidth; x++ {
			sx := bounds.Min.X + int(float64(x)/float64(targetWidth)*float64(width))
			r, g, b, a := img.At(sx, sy).RGBA()
			if a == 0 {
				buf.WriteByte(' ')
				continue
			}
			gray := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
			idx := int(gray * float64(len(chars)-1))
			buf.WriteByte(chars[idx])
		}
		buf.WriteByte('\n')
	}

	return buf.String()
}
