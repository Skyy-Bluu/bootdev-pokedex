package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	internals "github.com/skyy-bluu/internals"
)

type config struct {
	Next                string
	Prev                string
	Mapb                bool
	onFirstPage         bool
	ExploreLocationArea string
	CatchPokemon        string
	InspectPokemon      string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type LocationAreas struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type SpecificLocationAreaResponse struct {
	PokemonEncounters []PokemonEncounters `json:"pokemon_encounters"`
}

type PokemonEncounters struct {
	PokemonEncounter PokemonEncounter `json:"pokemon"`
}

type PokemonEncounter struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type LocationAreaResponse struct {
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Result   []LocationAreas `json:"results"`
}

type Pokemon struct {
	Name           string
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []Stat `json:"stats"`
	Types          []Type `json:"types"`
}

type Stat struct {
	Value int      `json:"base_stat"`
	Info  StatInfo `json:"stat"`
}

type StatInfo struct {
	Name string `json:"name"`
}

type Type struct {
	Info TypeInfo `json:"type"`
}

type TypeInfo struct {
	Name string `json:"name"`
}

func cleanInput(text string) []string {
	stringSlice := []string{}

	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	stringSlice = strings.Split(text, " ")
	return stringSlice
}

var commands map[string]cliCommand
var cache internals.Cache
var pokedex = make(map[string]Pokemon)

func init() {
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    exitCommand,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    helpCommand,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 location areas in the Pokemon world",
			callback:    mapCommand,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous names of 20 location areas in the Pokemon world",
			callback:    mapCommand,
		},
		"explore": {
			name:        "explore",
			description: "Displays a list pokemon available in that location",
			callback:    exploreCommand,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch the chosen Pokemon",
			callback:    catchCommand,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a Pokemon in your Pokedex",
			callback:    inspectCommand,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all the Pokemon in your pokedex",
			callback:    pokedexCommand,
		},
	}
}

func pokedexCommand(config *config) error {
	fmt.Println("Your Pokedex:")

	for pokemon := range pokedex {
		fmt.Printf("- %s \n", pokedex[pokemon].Name)
	}
	return nil
}

func inspectCommand(config *config) error {

	val, ok := pokedex[config.InspectPokemon]

	if !ok {
		return fmt.Errorf("you have not caught that pokemon")
	} else {
		fmt.Println("Name: ", val.Name)
		fmt.Println("Height: ", val.Height)
		fmt.Println("Weight: ", val.Weight)
		fmt.Println("Stats: ")
		for _, stat := range val.Stats {
			fmt.Printf("-%s: %v\n", stat.Info.Name, stat.Value)
		}
		fmt.Println("Types: ")

		for _, t := range val.Types {
			fmt.Printf("- %s\n", t.Info.Name)
		}
	}
	return nil
}

func catchCommand(config *config) error {

	fmt.Printf("Throwing a Pokeball at %s... \n", config.CatchPokemon)
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", config.CatchPokemon)
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	var pokemon Pokemon

	if err := json.Unmarshal(data, &pokemon); err != nil {
		return err
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	chance := (r.Intn(608) - pokemon.BaseExperience)

	pokemon.Name = config.CatchPokemon

	if chance > 0 {
		pokedex[config.CatchPokemon] = pokemon

		fmt.Printf("%s was caught \n", config.CatchPokemon)
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		return fmt.Errorf("%s escaped you loser HA! \n", config.CatchPokemon)
	}

	//fmt.Println(pokedex)
	return nil
}

func exploreCommand(config *config) error {
	fmt.Printf("Exploring %s...\n", config.ExploreLocationArea)
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", config.ExploreLocationArea)

	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	var specificLocationAreaResponse SpecificLocationAreaResponse

	if err := json.Unmarshal(data, &specificLocationAreaResponse); err != nil {
		return err
	}

	fmt.Println("List of Pokemon in this area:")
	for _, encounter := range specificLocationAreaResponse.PokemonEncounters {
		fmt.Println("- ", encounter.PokemonEncounter.Name)
	}

	return nil
}

func mapCommand(config *config) error {
	var url string

	switch {
	case config.onFirstPage && config.Mapb:
		config.Mapb = false
		fmt.Println("You're one the first page")
		return nil
	case config.Mapb:
		url = config.Prev
		config.Mapb = false
	case config.Next != "":
		url = config.Next
	default:
		url = "https://pokeapi.co/api/v2/location-area/"
	}

	data, ok := cache.Get(url)

	if !ok {
		fmt.Println("New Network Request....")
		res, err := http.Get(url)
		if err != nil {
			return err
		}

		defer res.Body.Close()

		data, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		cache.Add(url, data)
	} else {
		fmt.Println("Using cached data....")
	}

	var locatonAreas LocationAreaResponse

	if err := json.Unmarshal(data, &locatonAreas); err != nil {
		return err
	}

	config.Next = locatonAreas.Next
	config.Prev = locatonAreas.Previous

	if locatonAreas.Previous == "" {
		config.onFirstPage = true
	} else {
		config.onFirstPage = false
	}

	for _, locationName := range locatonAreas.Result {
		fmt.Println(locationName.Name)
	}

	return nil
}

func helpCommand(config *config) error {
	fmt.Println(`Welcome to the Pokedex!
Usage:`)
	for _, command := range commands {
		fmt.Println(command.description)
	}
	return nil
}

func exitCommand(config *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func main() {
	config := config{}
	scanner := bufio.NewScanner(os.Stdin)
	cache = internals.NewCache(time.Second * 20)

	for {
		fmt.Print("Pokedex > ")

		scanned := scanner.Scan()

		if scanned {
			text := scanner.Text()
			commandText := cleanInput(text)
			command, ok := commands[commandText[0]]

			if ok {
				if command.name == "mapb" {
					config.Mapb = true
				}
				if command.name == "explore" {
					config.ExploreLocationArea = commandText[1]
				}
				if command.name == "catch" {
					config.CatchPokemon = commandText[1]
				}
				if command.name == "inspect" {
					config.InspectPokemon = commandText[1]
				}
				err := command.callback(&config)

				if err != nil {
					fmt.Println("Oops:", err)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
