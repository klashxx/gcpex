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
	Args     []string
	Success  bool
	Pid      int
	OutBytes []byte
	Duration int
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
	var e Execution

	for c := range commands {
		path, err := exec.LookPath(c.Cmd)
		if err != nil {
			log.Println("Error -> Command:", c.Cmd, "Args:", c.Args, "Error:", err)
		} else {

			e.Path = path
			e.Cmd = c.Cmd
			e.Args = c.Args

			cmd := exec.Command(c.Cmd, c.Args...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Println("Error -> Command:", e.Cmd, "Args:", e.Args, "Error:", err)
			} else {
				err = cmd.Start()
				if err != nil {
					log.Println("Error -> Command:", e.Cmd, "Args:", e.Args, "Error:", err)
				} else {
					start := time.Now()
					e.Pid = cmd.Process.Pid

					log.Println("Start -> PID:", e.Pid, "Command:", e.Cmd, "Args:", e.Args)

					_, err = bufio.NewReader(stdout).Read(e.OutBytes)
					if err != nil {
						log.Println("Error -> Command:", e.Cmd, "Args:", e.Args, "Error:", err)
					} else {
						cmd.Wait()
						e.Duration = int(time.Since(start).Seconds())
						e.Success = cmd.ProcessState.Success()
						log.Println("End   -> PID:", e.Pid, "Command:", e.Cmd, "Args:", e.Args, "Duration", e.Duration)
					}
				}
			}
		}

		select {
		case executions <- e:
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
