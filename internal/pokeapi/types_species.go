package pokeapi

type NamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonSpeciesResponse struct {
	GrowthRate     NamedAPIResource `json:"growth_rate"`
	EvolutionChain struct {
		URL string `json:"url"`
	} `json:"evolution_chain"`
}

type GrowthRateResponse struct {
	Levels []struct {
		Level      int `json:"level"`
		Experience int `json:"experience"`
	} `json:"levels"`
}

type EvolutionChainResponse struct {
	Chain EvolutionChainLink `json:"chain"`
}

type EvolutionChainLink struct {
	Species          NamedAPIResource     `json:"species"`
	EvolvesTo        []EvolutionChainLink `json:"evolves_to"`
	EvolutionDetails []EvolutionDetail    `json:"evolution_details"`
}

type EvolutionDetail struct {
	MinLevel *int `json:"min_level"`
}
