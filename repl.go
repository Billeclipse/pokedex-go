package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"pokedex-go/internal/pokecache"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	NextLocationAreaURL     string
	PreviousLocationAreaURL string
	Cache                   *pokecache.Cache
}

type locationAreaResponse struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

const locationAreaURL = "https://pokeapi.co/api/v2/location-area"

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
	}
}

var commandOrder = []string{"help", "exit", "map", "mapb"}

func startREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	cfg := config{
		NextLocationAreaURL: locationAreaURL,
		Cache:               pokecache.NewCache(5 * time.Second),
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

		if err := command.callback(&cfg); err != nil {
			fmt.Println(err)
		}
	}
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
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

func commandMap(cfg *config) error {
	return fetchAndPrintLocationAreas(cfg, cfg.NextLocationAreaURL)
}

func commandMapb(cfg *config) error {
	if cfg.PreviousLocationAreaURL == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	return fetchAndPrintLocationAreas(cfg, cfg.PreviousLocationAreaURL)
}

func fetchAndPrintLocationAreas(cfg *config, url string) error {
	if cfg.Cache == nil {
		cfg.Cache = pokecache.NewCache(5 * time.Second)
	}

	body, ok := cfg.Cache.Get(url)
	if !ok {
		res, err := httpClient.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 {
			return fmt.Errorf("request failed with status %s", res.Status)
		}

		body, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		cfg.Cache.Add(url, body)
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

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(strings.TrimSpace(text)))
}
