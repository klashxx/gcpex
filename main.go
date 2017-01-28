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
	Error    error
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
		e.Cmd = c.Cmd
		e.Args = c.Args
		path, err := exec.LookPath(c.Cmd)
		if err != nil {
			e.Error = err
			executions <- e
			return
		}
		e.Path = path

		cmd := exec.Command(e.Cmd, e.Args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			e.Error = err
			executions <- e
			return
		}

		err = cmd.Start()
		if err != nil {
			e.Error = err
			executions <- e
			return
		}
		start := time.Now()
		e.Pid = cmd.Process.Pid

		log.Println("Start -> PID:", e.Pid, "Command:", e.Cmd, "Args:", e.Args, e.Error)

		_, err = bufio.NewReader(stdout).Read(e.OutBytes)
		if err != nil {
			e.Error = err
			executions <- e
			return
		}
		cmd.Wait()
		log.Println(e.Cmd, e.Error)
		e.Duration = int(time.Since(start).Seconds())
		e.Success = cmd.ProcessState.Success()

	}

	select {
	case executions <- e:
	case <-done:
		return
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

	for e := range executions {
		log.Println("End   -> PID:", e.Pid, "Command:", e.Cmd, "Args:", e.Args, "Duration", e.Duration, "Error", e.Error)
	}

	if err := <-errc; err != nil {
		log.Println(err)
	}
}
