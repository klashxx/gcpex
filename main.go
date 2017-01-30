package main

import (
	"bytes"
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
	Duration int
	Error    string
	Stdout   bytes.Buffer
	Stderr   bytes.Buffer
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
	execFile    string
	numRoutines int
)

func init() {
	flag.StringVar(&execFile, "exec", "", "cmd JSON file. [mandatory]")
	flag.IntVar(&numRoutines, "routines", 5, "max parallel execution routines")
}

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

func commandDigester(done <-chan struct{}, commands <-chan Command, executions chan<- Execution) {
	var e Execution

	for c := range commands {
		path, err := exec.LookPath(c.Cmd)
		if err != nil {
			e.Error = err.Error()
		} else {
			e.Path = path
			e.Cmd = c.Cmd
			e.Args = c.Args

			cmd := exec.Command(e.Cmd, e.Args...)

			err = cmd.Start()
			if err != nil {
				e.Error = err.Error()
			} else {
				start := time.Now()
				e.Pid = cmd.Process.Pid
				log.Println("Start -> PID:", e.Pid, "Command:", e.Cmd, "Args:", e.Args)

				cmd.Wait()
				e.Duration = int(time.Since(start).Seconds())
				e.Success = cmd.ProcessState.Success()
			}
		}

		select {
		case executions <- e:
		case <-done:
			return
		}
	}
}

func controller(c Commands) bool {
	success := true

	done := make(chan struct{})
	defer close(done)

	commands, errc := dispatchCommands(done, c)

	var wg sync.WaitGroup
	wg.Add(numRoutines)

	executions := make(chan Execution)

	for i := 0; i < numRoutines; i++ {
		go func() {
			commandDigester(done, commands, executions)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(executions)
	}()

	for e := range executions {
		log.Println(e)
		log.Println(e.Stdout.String())
		if !e.Success {
			success = false
		}
	}

	if err := <-errc; err != nil {
		log.Println(err)
		success = false
	}

	return success
}

func main() {
	flag.Parse()
	if execFile == "" {
		flag.PrintDefaults()
		os.Exit(5)
	}

	c, err := deserializeJSON(execFile)
	if err != nil {
		log.Fatal(err)
	}

	_ = controller(c)

}
