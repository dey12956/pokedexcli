package pokeapi

func (c *Client) ListLocations(pageURL *string) (Response, error) {
	url := baseURL + "/location-area"
	if pageURL != nil {
		url = *pageURL
	}

	locationsResp := Response{}
	if err := c.getResource(url, &locationsResp); err != nil {
		return Response{}, err
	}

	return locationsResp, nil
}
