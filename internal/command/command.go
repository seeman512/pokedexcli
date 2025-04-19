package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	plogger "pokedexcli/internal/logger"
	"pokedexcli/internal/pokecache"
	"strings"
	"time"
)

const BASE_URL = "https://pokeapi.co/api/v2/"
const LOCATION_AREA_URL = BASE_URL + "/location-area"
const NONE = "NONE"

var logger plogger.Logger = plogger.New(plogger.TRACE, os.Stderr, "COMMAND: ")
var cache = pokecache.NewCache(60 * time.Second)

type commandCb func(*Config, string) error

type cliCommand struct {
	Name        string
	Description string
	Callback    commandCb
}

type Config struct {
	Next     string
	Previous string
}

type LocationAreasResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationAreaPokemonResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

var commands map[string]cliCommand = map[string]cliCommand{
	"exit": {
		Name:        "exit",
		Description: "Exit the Pokedex",
		Callback:    commandExit,
	},
	"help": {
		Name:        "help",
		Description: "Display a help message",
		Callback:    commandHelp,
	},
	"map": {
		Name:        "map",
		Description: "Display Next Locations",
		Callback:    commandMap,
	},
	"mapb": {
		Name:        "mapb",
		Description: "Display Previous Locations",
		Callback:    commandMapB,
	},
	"explore": {
		Name:        "explore",
		Description: "Display location pokemons",
		Callback:    commandMapB,
	},
}

var usage []string = []string{}

func commandExit(*Config, string) error {
	println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(*Config, string) error {
	help := `
Welcome to the Pokedex!
Usage:

`
	println(help)
	println(strings.Join(usage, "\n"))

	return nil
}

func commandMap(c *Config, s string) error {
	if c.Next == NONE {
		return errors.New("limit max pages")
	}

	if c.Next == "" {
		c.Next = LOCATION_AREA_URL
	}

	prev, next, locations, err := getLocationAreas(c.Next)
	if err != nil {
		return err
	}

	if prev == nil {
		c.Previous = NONE
	} else {
		c.Previous = *prev
	}

	if next == nil {
		c.Next = NONE
	} else {
		c.Next = *next
	}

	for _, l := range locations {
		println(l)
	}

	return nil
}

func commandMapB(c *Config, s string) error {
	if c.Previous == NONE || c.Previous == "" {
		return errors.New("wrong page")
	}

	prev, next, locations, err := getLocationAreas(c.Previous)
	if err != nil {
		return err
	}

	if prev == nil {
		c.Previous = NONE
	} else {
		c.Previous = *prev
	}

	if next == nil {
		c.Next = NONE
	} else {
		c.Next = *next
	}

	for _, l := range locations {
		println(l)
	}

	return nil
}

func getLocationAreas(url string) (*string, *string, []string, error) {
	var prev *string = nil
	var next *string = nil

	body, err := getBody(url)
	if err != nil {
		return prev, next, nil, err
	}

	response := LocationAreasResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return prev, next, nil, fmt.Errorf("unmarshall error %w", err)
	}

	locations := []string{}
	for _, r := range response.Results {
		locations = append(locations, r.Name)
	}

	return response.Previous, response.Next, locations, nil
}

func getBody(url string) ([]byte, error) {
	body, ok := cache.Get(url)
	if ok {
		return body, nil
	}

	logger.Trace("Fetch: " + url)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err = io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("url %s error code %d; %body", url, res.StatusCode, body)
	}

	if err != nil {
		return nil, err
	}

	cache.Add(url, body)
	return body, nil
}

func Init() {
	for _, cmd := range commands {
		usage = append(usage, fmt.Sprintf("%s: %s", cmd.Name, cmd.Description))
	}
}

func GetCommands() map[string]cliCommand {
	return commands
}
