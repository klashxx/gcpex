package main

import "flag"

var (
	author      = "klashxx@gmail.com"
	execFile    = flag.String("exec", "", "cmd JSON file. [obligatory]")
	numRoutines = flag.Int("routines", 5, "max parallel execution routines")
)

func main() {
	flag.Parse()

	if *execFile == "" {
		flag.PrintDefaults()
	}

}
