package pokeapi

type MoveResponse struct {
	Name     string `json:"name"`
	Power    *int   `json:"power"`
	Accuracy *int   `json:"accuracy"`
	Priority int    `json:"priority"`
	Type     struct {
		Name string `json:"name"`
	} `json:"type"`
}
