package pokeapi

import "net/url"

func (c *Client) GetPokemonSpecies(name string) (PokemonSpeciesResponse, error) {
	resourceURL := baseURL + "/pokemon-species/" + url.PathEscape(name)
	resp := PokemonSpeciesResponse{}
	if err := c.getResource(resourceURL, &resp); err != nil {
		return PokemonSpeciesResponse{}, err
	}
	return resp, nil
}

func (c *Client) GetGrowthRate(resourceURL string) (GrowthRateResponse, error) {
	resp := GrowthRateResponse{}
	if err := c.getResource(resourceURL, &resp); err != nil {
		return GrowthRateResponse{}, err
	}
	return resp, nil
}

func (c *Client) GetEvolutionChain(resourceURL string) (EvolutionChainResponse, error) {
	resp := EvolutionChainResponse{}
	if err := c.getResource(resourceURL, &resp); err != nil {
		return EvolutionChainResponse{}, err
	}
	return resp, nil
}
