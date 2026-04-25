# Pokedex Go

A command-line Pokedex built in Go. The app runs as an interactive REPL and uses the [PokeAPI](https://pokeapi.co/) to explore Pokemon locations, discover Pokemon encounters, catch Pokemon, inspect caught Pokemon, and list your personal Pokedex.

## Features

- Interactive `Pokedex >` command prompt
- `help` and `exit` commands
- Paginated location browsing with `map` and `mapb`
- Location exploration with `explore <location-area>`
- PokeAPI response caching for faster repeated navigation
- Pokemon catching with randomized catch chance based on base experience
- In-memory Pokedex of caught Pokemon
- Pokemon inspection for caught Pokemon, including height, weight, stats, and types
- Unit tests for input cleaning, caching, API fetch behavior, inspect output, and Pokedex listing

## Commands

```text
help                         Displays available commands
exit                         Exits the Pokedex
map                          Shows the next 20 location areas
mapb                         Shows the previous 20 location areas
explore <location-area>      Lists Pokemon found in a location area
catch <pokemon>              Attempts to catch a Pokemon
inspect <pokemon>            Shows details for a caught Pokemon
pokedex                      Lists all caught Pokemon
```

## Example Session

```text
Pokedex > map
canalave-city-area
eterna-city-area
pastoria-city-area
...
Pokedex > explore canalave-city-area
Exploring canalave-city-area...
Found Pokemon:
 - tentacool
 - tentacruel
 - staryu
 - magikarp
Pokedex > catch pidgey
Throwing a Pokeball at pidgey...
pidgey was caught!
You may now inspect it with the inspect command.
Pokedex > inspect pidgey
Name: pidgey
Height: 3
Weight: 18
Stats:
  -hp: 40
  -attack: 45
  -defense: 40
  -special-attack: 35
  -special-defense: 35
  -speed: 56
Types:
  - normal
  - flying
Pokedex > pokedex
Your Pokedex:
 - pidgey
```

## Requirements

- Go 1.26.2 or newer
- Network access to `https://pokeapi.co`

## Getting Started

```bash
go run .
```

Build the binary:

```bash
go build
./pokedex-go
```

Run tests:

```bash
go test ./...
```

## Project Structure

```text
.
├── main.go
├── repl.go
├── repl_test.go
├── internal/
│   └── pokecache/
│       ├── pokecache.go
│       └── pokecache_test.go
├── docs/
│   └── lesson_notes.json
├── output/pdf/
│   └── pokedex-go-lessons.pdf
└── scripts/
    └── generate_lesson_pdf.py
```

## Notes

The app stores caught Pokemon in memory, so your Pokedex resets when the program exits. API responses are cached in memory for a short interval to make repeated navigation faster while keeping data fresh.
