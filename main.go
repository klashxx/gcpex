package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var (
	author      = "klashxx@gmail.com"
	execFile    = flag.String("exec", "", "cmd JSON file. [obligatory]")
	numRoutines = flag.Int("routines", 5, "max parallel execution routines")
)

func deserializeJSON(execFile string) ([]byte, error) {
	rawJSON, err := ioutil.ReadFile(execFile)
	if err != nil {
		return rawJSON, err
	}
	return rawJSON, nil
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
}
