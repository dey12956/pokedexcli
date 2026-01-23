package pokeapi

import "net/url"

func (c *Client) ListPokemon(area string) (PokemonResponse, error) {
	url := baseURL + "/location-area/" + url.PathEscape(area)

	pokemonResp := PokemonResponse{}
	if err := c.getResource(url, &pokemonResp); err != nil {
		return PokemonResponse{}, err
	}

	return pokemonResp, nil

}
