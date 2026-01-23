package pokeapi

import "net/url"

func (c *Client) CatchPokemon(name string) (CatchPokemonResponse, error) {
	return c.getPokemon(name)
}

func (c *Client) GetPokemon(name string) (CatchPokemonResponse, error) {
	return c.getPokemon(name)
}

func (c *Client) getPokemon(name string) (CatchPokemonResponse, error) {
	url := baseURL + "/pokemon/" + url.PathEscape(name)
	catchPokeResp := CatchPokemonResponse{}
	if err := c.getResource(url, &catchPokeResp); err != nil {
		return CatchPokemonResponse{}, err
	}

	return catchPokeResp, nil
}
