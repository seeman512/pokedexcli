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
	config := command.NewConfig()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		parts := cleanInput(line)
		cmd, arg := parts[0], ""

		if len(parts) >= 2 {
			arg = parts[1]
		}

		command, ok := commands[cmd]
		if !ok {
			println("uknown command")
			continue
		}

		err := command.Callback(config, arg)
		if err != nil {
			fmt.Printf("Command error %v\n", err)
		}
	}
}
