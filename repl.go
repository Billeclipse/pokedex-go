package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"pokedex-go/internal/pokecache"
	"sort"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type config struct {
	NextLocationAreaURL     string
	PreviousLocationAreaURL string
	Cache                   *pokecache.Cache
	Pokedex                 map[string]Pokemon
}

type locationAreaResponse struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

const locationAreaURL = "https://pokeapi.co/api/v2/location-area"
const pokemonURL = "https://pokeapi.co/api/v2/pokemon"

var httpClient = http.Client{
	Timeout: 10 * time.Second,
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Displays Pokemon in a location area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempts to catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Displays details for caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Displays caught Pokemon",
			callback:    commandPokedex,
		},
	}
}

var commandOrder = []string{"help", "exit", "map", "mapb", "explore", "catch", "inspect", "pokedex"}

func startREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	cfg := config{
		NextLocationAreaURL: locationAreaURL,
		Cache:               pokecache.NewCache(5 * time.Second),
		Pokedex:             make(map[string]Pokemon),
	}

	for {
		fmt.Print("Pokedex > ")

		if !scanner.Scan() {
			break
		}

		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}

		command, ok := commands[words[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		if err := command.callback(&cfg, words[1:]); err != nil {
			fmt.Println(err)
		}
	}
}

func commandExit(cfg *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, args []string) error {
	commands := getCommands()

	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, name := range commandOrder {
		command := commands[name]
		fmt.Printf("%s: %s\n", command.name, command.description)
	}

	return nil
}

func commandMap(cfg *config, args []string) error {
	return fetchAndPrintLocationAreas(cfg, cfg.NextLocationAreaURL)
}

func commandMapb(cfg *config, args []string) error {
	if cfg.PreviousLocationAreaURL == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	return fetchAndPrintLocationAreas(cfg, cfg.PreviousLocationAreaURL)
}

func commandExplore(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a location area name")
	}

	areaName := args[0]
	return fetchAndPrintPokemonInLocation(cfg, fmt.Sprintf("%s/%s", locationAreaURL, areaName), areaName)
}

func commandCatch(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a pokemon name")
	}

	pokemonName := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	pokemon, err := fetchPokemon(cfg, fmt.Sprintf("%s/%s", pokemonURL, pokemonName))
	if err != nil {
		return err
	}

	if cfg.Pokedex == nil {
		cfg.Pokedex = make(map[string]Pokemon)
	}

	if wasCaught(pokemon) {
		cfg.Pokedex[pokemon.Name] = pokemon
		fmt.Printf("%s was caught!\n", pokemon.Name)
		fmt.Println("You may now inspect it with the inspect command.")
		return nil
	}

	fmt.Printf("%s escaped!\n", pokemon.Name)
	return nil
}

func commandInspect(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a pokemon name")
	}

	pokemon, ok := cfg.Pokedex[args[0]]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")

	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, pokemonType := range pokemon.Types {
		fmt.Printf("  - %s\n", pokemonType.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *config, args []string) error {
	fmt.Println("Your Pokedex:")

	names := make([]string, 0, len(cfg.Pokedex))
	for name := range cfg.Pokedex {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf(" - %s\n", name)
	}

	return nil
}

func fetchBody(cfg *config, url string) ([]byte, error) {
	if cfg.Cache == nil {
		cfg.Cache = pokecache.NewCache(5 * time.Second)
	}

	body, ok := cfg.Cache.Get(url)
	if ok {
		return body, nil
	}

	res, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %s", res.Status)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	cfg.Cache.Add(url, body)

	return body, nil
}

func fetchAndPrintLocationAreas(cfg *config, url string) error {
	body, err := fetchBody(cfg, url)
	if err != nil {
		return err
	}

	areas := locationAreaResponse{}
	if err := json.Unmarshal(body, &areas); err != nil {
		return err
	}

	cfg.NextLocationAreaURL = areas.Next
	cfg.PreviousLocationAreaURL = areas.Previous

	for _, area := range areas.Results {
		fmt.Println(area.Name)
	}

	return nil
}

type locationAreaDetail struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func fetchAndPrintPokemonInLocation(cfg *config, url string, areaName string) error {
	body, err := fetchBody(cfg, url)
	if err != nil {
		return err
	}

	area := locationAreaDetail{}
	if err := json.Unmarshal(body, &area); err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", areaName)
	fmt.Println("Found Pokemon:")

	for _, encounter := range area.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func fetchPokemon(cfg *config, url string) (Pokemon, error) {
	body, err := fetchBody(cfg, url)
	if err != nil {
		return Pokemon{}, err
	}

	pokemon := Pokemon{}
	if err := json.Unmarshal(body, &pokemon); err != nil {
		return Pokemon{}, err
	}

	return pokemon, nil
}

func wasCaught(pokemon Pokemon) bool {
	catchTarget := pokemon.BaseExperience
	if catchTarget < 1 {
		catchTarget = 1
	}

	return rand.Intn(catchTarget) < 50
}

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(strings.TrimSpace(text)))
}
