package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) ListLocations(pageURL *string) (Response, error) {
	url := baseURL + "/location-area"
	if pageURL != nil {
		url = *pageURL
	}

	if data, exists := c.cache.Get(url); exists {
		locationResp := Response{}
		err := json.Unmarshal(data, &locationResp)
		if err != nil {
			return Response{}, err
		}
		return locationResp, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Response{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Response{}, fmt.Errorf("pokeapi error: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	c.cache.Add(url, data)

	locationsResp := Response{}
	err = json.Unmarshal(data, &locationsResp)
	if err != nil {
		return Response{}, err
	}

	return locationsResp, nil
}
