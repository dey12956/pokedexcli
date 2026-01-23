package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) ListPokemon(area string) (PokemonResponse, error) {
	url := baseURL + "/location-area/" + url.PathEscape(area)

	if data, exists := c.cache.Get(url); exists {
		pokemonResp := PokemonResponse{}
		err := json.Unmarshal(data, &pokemonResp)
		if err != nil {
			c.cache.Delete(url)
		} else {
			return pokemonResp, nil
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return PokemonResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PokemonResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return PokemonResponse{}, fmt.Errorf("pokeapi error: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return PokemonResponse{}, err
	}

	c.cache.Add(url, data)

	pokemonResp := PokemonResponse{}
	err = json.Unmarshal(data, &pokemonResp)
	if err != nil {
		return PokemonResponse{}, err
	}

	return pokemonResp, nil

}
