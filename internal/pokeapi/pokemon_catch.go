package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) CatchPokemon(name string) (CatchPokemonResponse, error) {
	url := baseURL + "/pokemon/" + url.PathEscape(name)

	if data, exists := c.cache.Get(url); exists {
		catchPokeResp := CatchPokemonResponse{}
		err := json.Unmarshal(data, &catchPokeResp)
		if err != nil {
			c.cache.Delete(url)
		} else {
			return catchPokeResp, nil
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return CatchPokemonResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CatchPokemonResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return CatchPokemonResponse{}, fmt.Errorf("pokeapi error: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return CatchPokemonResponse{}, err
	}

	c.cache.Add(url, data)

	catchPokeResp := CatchPokemonResponse{}
	err = json.Unmarshal(data, &catchPokeResp)
	if err != nil {
		return CatchPokemonResponse{}, err
	}

	return catchPokeResp, nil
}
