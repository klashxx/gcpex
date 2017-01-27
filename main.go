package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

type Command struct {
	Profile string   `json:"profile"`
	Cmd     string   `json:"cmd"`
	Args    []string `json:"args"`
	Log     string   `json:"log"`
}

type Commands []Command

var (
	author      = "klashxx@gmail.com"
	execFile    = flag.String("exec", "", "cmd JSON file. [mandatory]")
	numRoutines = flag.Int("routines", 5, "max parallel execution routines")
)

func deserializeJSON(execFile string) (c Commands, err error) {
	rawJSON, err := ioutil.ReadFile(execFile)
	if err != nil {
		return c, err
	}

	err = json.Unmarshal(rawJSON, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func dispatchCommands(done <-chan struct{}, c Commands) (<-chan Command, <-chan error) {
	commands := make(chan Command)
	errc := make(chan error, 1)

	go func() {
		defer close(commands)

		errc <- func() error {
			for _, p := range c {
				select {
				case commands <- p:
				case <-done:
					return errors.New("dispatch canceled")
				}
			}
			return nil
		}()
	}()

	return commands, errc
}

func main() {
	flag.Parse()
	if *execFile == "" {
		flag.PrintDefaults()
		os.Exit(5)
	}

	c, err := deserializeJSON(*execFile)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	defer close(done)

	_, _ = dispatchCommands(done, c)
}
