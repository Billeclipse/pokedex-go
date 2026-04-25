package main

import (
	"net/http"
	"net/http/httptest"
	"pokedex-go/internal/pokecache"
	"sync/atomic"
	"testing"
	"time"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "\tSquirtle\nMewtwo   Eevee\t",
			expected: []string{"squirtle", "mewtwo", "eevee"},
		},
		{
			input:    "   ",
			expected: []string{},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("cleanInput(%q) returned %d words, expected %d", c.input, len(actual), len(c.expected))
			continue
		}

		for i := range actual {
			if actual[i] != c.expected[i] {
				t.Errorf("cleanInput(%q)[%d] = %q, expected %q", c.input, i, actual[i], c.expected[i])
			}
		}
	}
}

func TestFetchAndPrintLocationAreasUsesCache(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"next": "https://example.com/next",
			"previous": null,
			"results": [
				{"name": "test-area", "url": "https://example.com/location-area/1/"}
			]
		}`))
	}))
	defer server.Close()

	oldClient := httpClient
	httpClient = *server.Client()
	defer func() {
		httpClient = oldClient
	}()

	cfg := config{
		Cache: pokecache.NewCache(time.Minute),
	}

	if err := fetchAndPrintLocationAreas(&cfg, server.URL); err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}

	if err := fetchAndPrintLocationAreas(&cfg, server.URL); err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}

	if requests != 1 {
		t.Fatalf("expected 1 HTTP request, got %d", requests)
	}
}

func TestFetchAndPrintPokemonInLocationUsesCache(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"pokemon_encounters": [
				{
					"pokemon": {
						"name": "pikachu",
						"url": "https://example.com/pokemon/25/"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	oldClient := httpClient
	httpClient = *server.Client()
	defer func() {
		httpClient = oldClient
	}()

	cfg := config{
		Cache: pokecache.NewCache(time.Minute),
	}

	if err := fetchAndPrintPokemonInLocation(&cfg, server.URL, "test-area"); err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}

	if err := fetchAndPrintPokemonInLocation(&cfg, server.URL, "test-area"); err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}

	if requests != 1 {
		t.Fatalf("expected 1 HTTP request, got %d", requests)
	}
}
