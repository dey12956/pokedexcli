package pokeapi

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/dey12956/pokedexcli/internal/pokecache"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestListPokemonUsesEscapedURLInCache(t *testing.T) {
	area := "kanto-route 1"
	expectedURL := baseURL + "/location-area/" + url.PathEscape(area)

	cache := pokecache.NewCache(time.Second, time.Second)
	t.Cleanup(cache.Close)
	cache.Add(expectedURL, []byte("{}"))

	called := false
	client := Client{
		httpClient: http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			return nil, fmt.Errorf("unexpected http request: %s", req.URL.String())
		})},
		cache: cache,
	}

	_, err := client.ListPokemon(area)
	if err != nil {
		t.Fatalf("expected cached response, got error: %v", err)
	}
	if called {
		t.Fatalf("expected no HTTP call when cache is primed")
	}
}

func TestCatchPokemonUsesEscapedURLInCache(t *testing.T) {
	name := "mr mime"
	expectedURL := baseURL + "/pokemon/" + url.PathEscape(name)

	cache := pokecache.NewCache(time.Second, time.Second)
	t.Cleanup(cache.Close)
	cache.Add(expectedURL, []byte("{}"))

	called := false
	client := Client{
		httpClient: http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			return nil, fmt.Errorf("unexpected http request: %s", req.URL.String())
		})},
		cache: cache,
	}

	_, err := client.CatchPokemon(name)
	if err != nil {
		t.Fatalf("expected cached response, got error: %v", err)
	}
	if called {
		t.Fatalf("expected no HTTP call when cache is primed")
	}
}
