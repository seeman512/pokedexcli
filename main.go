package main

import (
	"bufio"
	"fmt"
	"os"
	"pokedexcli/internal/command"
)

func main() {
	command.Init()
	commands := command.GetCommands()
	config := &command.Config{}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		command, ok := commands[line]
		if !ok {
			println("uknown command")
			continue
		}

		err := command.Callback(config, "")
		if err != nil {
			fmt.Printf("Command error %v\n", err)
		}
	}
}
