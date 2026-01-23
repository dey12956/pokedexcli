package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) getResource(url string, target any) error {
	if data, exists := c.cache.Get(url); exists {
		if err := json.Unmarshal(data, target); err == nil {
			return nil
		}
		c.cache.Delete(url)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("pokeapi error: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.cache.Add(url, data)

	if err := json.Unmarshal(data, target); err != nil {
		return err
	}

	return nil
}
