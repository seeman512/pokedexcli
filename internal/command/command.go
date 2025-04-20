package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	plogger "pokedexcli/internal/logger"
	"pokedexcli/internal/pokecache"
	"strings"
	"time"
)

const BASE_URL = "https://pokeapi.co/api/v2/"
const LOCATION_AREA_URL = BASE_URL + "/location-area"
const POKEMON_URL = BASE_URL + "/pokemon"
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
	Pokedex  map[string]Pokemon
}

type LocationAreas struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationAreaPokemon struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name           string `json:"name"`
	Weight         int    `json:"weight"`
	Height         int    `json:"height"`
	BaseExperience int    `json:"base_experience"`
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
		Callback:    commandExplore,
	},
	"catch": {
		Name:        "catch",
		Description: "Catch pokemon",
		Callback:    commandCatch,
	},
	"inspect": {
		Name:        "inspect",
		Description: "Inspect pokemon",
		Callback:    commandInspect,
	},
}

var usage []string = []string{}

func (p Pokemon) getStats() map[string]int {
	stats := map[string]int{}

	for _, stat := range p.Stats {
		stats[stat.Stat.Name] = stat.BaseStat
	}

	return stats
}

func (p Pokemon) getTypes() []string {
	types := []string{}

	for _, t := range p.Types {
		types = append(types, t.Type.Name)
	}

	return types
}

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

func commandExplore(c *Config, area string) error {
	if area == "" {
		return errors.New("required area")
	}

	pokemons, err := getLocationPokemons(LOCATION_AREA_URL + "/" + area + "/")

	if err != nil {
		return err
	}

	if len(pokemons) == 0 {
		return errors.New("no pokemons found")
	}

	fmt.Println("Found Pokemon:")
	for _, pokemon := range pokemons {
		fmt.Printf("- %s\n", pokemon)
	}

	return nil
}

func isCatch(baseExperience int) bool {
	r := 2
	if baseExperience < 50 {
		r = 3
	} else if baseExperience < 75 {
		r = 4
	} else {
		r = 5
	}

	return rand.Intn(r) == 0
}

func commandCatch(c *Config, pokemon string) error {
	if pokemon == "" {
		return errors.New("required area")
	}

	if _, ok := c.Pokedex[pokemon]; ok {
		fmt.Printf("%s has been already catched!\n", pokemon)
		return nil
	}

	pokemonInst, err := getPokemon(POKEMON_URL + "/" + pokemon + "/")

	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon)
	if isCatch(pokemonInst.BaseExperience) {
		fmt.Printf("%s was caught!\n", pokemon)
		c.Pokedex[pokemon] = *pokemonInst
	} else {
		fmt.Printf("%s escaped!\n", pokemon)
	}

	return nil
}

func commandInspect(c *Config, pokemon string) error {
	pokemonInst, ok := c.Pokedex[pokemon]

	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	statsStr := ""
	for k, v := range pokemonInst.getStats() {
		statsStr += fmt.Sprintf("  -%s: %d\n", k, v)
	}

	typesStr := ""
	for _, t := range pokemonInst.getTypes() {
		typesStr += fmt.Sprintf("  -%s\n", t)
	}

	fmt.Printf(`
Name: %s
Height: %d
Weight: %d
Stats:
%s
Types:
%s`, pokemonInst.Name, pokemonInst.Height, pokemonInst.Weight, statsStr, typesStr)
	return nil
}

func getPokemon(url string) (*Pokemon, error) {
	body, err := getBody(url)
	if err != nil {
		return nil, err
	}

	response := Pokemon{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshall error %w", err)
	}

	return &response, nil
}

func getLocationPokemons(url string) ([]string, error) {
	body, err := getBody(url)
	if err != nil {
		return nil, err
	}

	response := LocationAreaPokemon{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("unmarshall error %w", err)
	}

	pokemons := []string{}
	for _, item := range response.PokemonEncounters {
		pokemons = append(pokemons, item.Pokemon.Name)
	}

	return pokemons, nil
}

func getLocationAreas(url string) (*string, *string, []string, error) {
	var prev *string = nil
	var next *string = nil

	body, err := getBody(url)
	if err != nil {
		return prev, next, nil, err
	}

	response := LocationAreas{}
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
		return nil, fmt.Errorf("url %s error code %d; %s", url, res.StatusCode, body)
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

func NewConfig() *Config {
	return &Config{
		Pokedex: make(map[string]Pokemon),
	}
}
