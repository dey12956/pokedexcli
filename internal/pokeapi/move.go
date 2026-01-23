package pokeapi

func (c *Client) GetMove(resourceURL string) (MoveResponse, error) {
	resp := MoveResponse{}
	if err := c.getResource(resourceURL, &resp); err != nil {
		return MoveResponse{}, err
	}
	return resp, nil
}
