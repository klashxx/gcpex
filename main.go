package main

import (
	"encoding/json"
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

func main() {
	flag.Parse()
	if *execFile == "" {
		flag.PrintDefaults()
		os.Exit(5)
	}

	_, err := deserializeJSON(*execFile)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	defer close(done)
}
