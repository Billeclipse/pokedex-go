package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"pokedex-go/internal/pokecache"
	"strings"
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

func TestFetchPokemonUsesCache(t *testing.T) {
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "pikachu",
			"base_experience": 112,
			"height": 4,
			"weight": 60,
			"stats": [
				{"base_stat": 35, "stat": {"name": "hp"}}
			],
			"types": [
				{"type": {"name": "electric"}}
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

	first, err := fetchPokemon(&cfg, server.URL)
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}

	second, err := fetchPokemon(&cfg, server.URL)
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}

	if first.Name != "pikachu" || second.Name != "pikachu" {
		t.Fatalf("expected pikachu responses, got %q and %q", first.Name, second.Name)
	}

	if requests != 1 {
		t.Fatalf("expected 1 HTTP request, got %d", requests)
	}
}

func TestWasCaughtAlwaysCatchesZeroBaseExperience(t *testing.T) {
	if !wasCaught(Pokemon{Name: "testmon", BaseExperience: 0}) {
		t.Fatalf("expected pokemon with zero base experience to be caught")
	}
}

func TestCommandInspectPrintsCaughtPokemon(t *testing.T) {
	cfg := config{
		Pokedex: map[string]Pokemon{
			"pidgey": {
				Name:   "pidgey",
				Height: 3,
				Weight: 18,
				Stats: []struct {
					BaseStat int `json:"base_stat"`
					Stat     struct {
						Name string `json:"name"`
					} `json:"stat"`
				}{
					{
						BaseStat: 40,
						Stat: struct {
							Name string `json:"name"`
						}{
							Name: "hp",
						},
					},
				},
				Types: []struct {
					Type struct {
						Name string `json:"name"`
					} `json:"type"`
				}{
					{
						Type: struct {
							Name string `json:"name"`
						}{
							Name: "normal",
						},
					},
				},
			},
		},
	}

	output := captureStdout(t, func() {
		if err := commandInspect(&cfg, []string{"pidgey"}); err != nil {
			t.Fatalf("inspect failed: %v", err)
		}
	})

	for _, expected := range []string{
		"Name: pidgey",
		"Height: 3",
		"Weight: 18",
		"Stats:",
		"  -hp: 40",
		"Types:",
		"  - normal",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestCommandInspectPrintsMissingPokemonMessage(t *testing.T) {
	cfg := config{
		Pokedex: map[string]Pokemon{},
	}

	output := captureStdout(t, func() {
		if err := commandInspect(&cfg, []string{"pidgey"}); err != nil {
			t.Fatalf("inspect failed: %v", err)
		}
	})

	if !strings.Contains(output, "you have not caught that pokemon") {
		t.Fatalf("expected missing pokemon message, got:\n%s", output)
	}
}

func TestCommandPokedexPrintsCaughtPokemon(t *testing.T) {
	cfg := config{
		Pokedex: map[string]Pokemon{
			"pidgey":   {Name: "pidgey"},
			"caterpie": {Name: "caterpie"},
		},
	}

	output := captureStdout(t, func() {
		if err := commandPokedex(&cfg, nil); err != nil {
			t.Fatalf("pokedex failed: %v", err)
		}
	})

	expected := "Your Pokedex:\n - caterpie\n - pidgey\n"
	if output != expected {
		t.Fatalf("expected output %q, got %q", expected, output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}

	os.Stdout = writer
	fn()
	_ = writer.Close()
	os.Stdout = oldStdout

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(reader); err != nil {
		t.Fatalf("read stdout failed: %v", err)
	}

	return buffer.String()
}
