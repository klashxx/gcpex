package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Execution struct {
	Cmd      string
	Path     string
	Success  bool
	Pid      int
	OutBytes []byte
}

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

func commandLauncher(done <-chan struct{}, commands <-chan Command, executions chan<- Execution) {
	var execution Execution

	for command := range commands {
		path, err := exec.LookPath(command.Cmd)
		if err != nil {
			log.Println("Error -> Command:", command.Cmd, "Args:", command.Args, "Error:", err)
		} else {

			execution.Path = path
			execution.Cmd = command.Cmd

			cmd := exec.Command(command.Cmd, command.Args...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Println("Error -> Command:", command.Cmd, "Args:", command.Args, "Error:", err)
			} else {

				err = cmd.Start()
				if err != nil {
					log.Println("Error -> Command:", command.Cmd, "Args:", command.Args, "Error:", err)
				} else {
					start := time.Now()
					execution.Pid = cmd.Process.Pid

					log.Println("Start -> PID:", execution.Pid, "Command:", command.Cmd, "Args:", command.Args)

					_, err = bufio.NewReader(stdout).Read(execution.OutBytes)
					if err != nil {
						log.Println("Error -> Command:", command.Cmd, "Args:", command.Args, "Error:", err)
					} else {
						cmd.Wait()
						duration := time.Since(start)
						log.Println("End   -> PID:", execution.Pid, "Command:", command.Cmd, "Args:", command.Args, "Duration", duration)
					}
				}
			}
		}

		select {
		case executions <- execution:
		case <-done:
			return
		}
	}
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

	commands, errc := dispatchCommands(done, c)

	var wg sync.WaitGroup
	wg.Add(*numRoutines)

	executions := make(chan Execution)

	for i := 0; i < *numRoutines; i++ {
		go func() {
			commandLauncher(done, commands, executions)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(executions)
	}()

	for ex := range executions {
		log.Println(ex)
	}

	if err := <-errc; err != nil {
		log.Println(err)
	}
}
