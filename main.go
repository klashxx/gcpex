package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type Execution struct {
	Cmd       string
	Path      string
	Env       []string
	Args      []string
	Success   bool
	Pid       int
	Duration  int
	Errors    []string
	Log       string
	Overwrite bool
}

type Executions []Execution

type Command struct {
	Cmd       string   `json:"cmd"`
	Args      []string `json:"args"`
	Env       []string `json:"env"`
	Log       string   `json:"log"`
	Overwrite bool     `json:"overwrite"`
}

type Commands []Command

const (
	author = "klashxx@gmail.com"
)

var (
	inJSON   string
	outJSON  string
	routines int
)

func init() {
	flag.StringVar(&inJSON, "in", "", "cmd JSON infile repo. [mandatory]")
	flag.StringVar(&outJSON, "out", "", "Respond JSON outfile.")
	flag.IntVar(&routines, "routines", 5, "max concurrent execution routines")
}

func isUsable(pathFile string, overWrite bool) error {

	_, err := os.Stat(pathFile)
	if os.IsExist(err) && !overWrite {
		return fmt.Errorf("file %s exists", pathFile)
	}

	_, err = os.Stat(path.Dir(pathFile))

	if os.IsNotExist(err) {
		return fmt.Errorf("%s: file base dir does not exists", pathFile)
	}

	if os.IsPermission(err) {
		return fmt.Errorf("%s: not enough permissions over base file directory", pathFile)
	}

	return nil
}

func streamToFile(l *os.File, outPipe io.ReadCloser, tag string) error {
	var err error
	var buf *bytes.Buffer
	var lock sync.Mutex
	block := bytes.Buffer{}

	if tag != "" {
		buf = bytes.NewBufferString(tag)
		buf.WriteTo(l)
	}

	bufC := 0
	end := make(chan error)
	go func() {
		var buf [1024]byte
		var err error
		var n int
		for err == nil {
			n, err = outPipe.Read(buf[:])
			if n > 0 {
				bufC++
				lock.Lock()
				block.Write(buf[:n])
				lock.Unlock()
			}
		}
		end <- err
	}()

	for err == nil {
		select {
		case err = <-end:
		default:
			lock.Lock()
			block.WriteTo(l)
			lock.Unlock()
		}
	}
	block.WriteTo(l)

	if tag != "" && bufC == 0 {
		buf = bytes.NewBufferString("<nil>\n")
		buf.WriteTo(l)
	}

	if err == io.EOF {
		return nil
	}
	return err
}

func deserializeJSON(inJSON string) (c Commands, err error) {
	rawJSON, err := ioutil.ReadFile(inJSON)
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
	for c := range commands {
		var e Execution
		var cmd *exec.Cmd
		var stdoutPipe io.ReadCloser
		var stderrPipe io.ReadCloser
		var l *os.File
		var start time.Time

		e.Cmd = c.Cmd
		e.Args = c.Args
		e.Env = c.Env
		e.Log = c.Log
		e.Overwrite = c.Overwrite

		path, err := exec.LookPath(c.Cmd)
		if err != nil {
			e.Errors = append(e.Errors, err.Error())
		}

		if len(e.Errors) == 0 {
			e.Path = filepath.Clean(path)
			if e.Log != "" {
				err = isUsable(e.Log, e.Overwrite)
				if err != nil {
					e.Errors = append(e.Errors, err.Error())
				} else {
					l, err = os.Create(e.Log)
					if err != nil {
						e.Errors = append(e.Errors, err.Error())
					}
					defer func(l *os.File) { l.Close() }(l)
				}
			}
		}

		if len(e.Errors) == 0 {
			cmd = exec.Command(e.Cmd, e.Args...)

			if e.Log != "" {
				stdoutPipe, err = cmd.StdoutPipe()
				if err != nil {
					e.Errors = append(e.Errors, err.Error())
				}
				stderrPipe, err = cmd.StderrPipe()
				if err != nil {
					e.Errors = append(e.Errors, err.Error())
				}
			}
		}

		if len(e.Errors) == 0 {
			err = cmd.Start()
			if err != nil {
				e.Errors = append(e.Errors, err.Error())
			}
		}

		if len(e.Errors) == 0 {
			start = time.Now()
			e.Pid = cmd.Process.Pid
			log.Println("Start -> Cmd:", e.Cmd, "Args:", e.Args, "PID:", e.Pid)

			if e.Log != "" {
				err = streamToFile(l, stdoutPipe, "STDOUT:\n=======\n\n")
				if err != nil {
					e.Errors = append(e.Errors, err.Error())
				}
				err = streamToFile(l, stderrPipe, "\nSTDERR:\n=======\n\n")
				if err != nil {
					e.Errors = append(e.Errors, err.Error())
				}
			}
		}

		if len(e.Errors) == 0 {
			err = cmd.Wait()
			if err != nil {
				e.Errors = append(e.Errors, err.Error())
			} else {
				e.Duration = int(time.Since(start).Seconds())
				e.Success = cmd.ProcessState.Success()
				log.Println("End   -> Cmd:", e.Cmd, "Args:", e.Args, "PID:", e.Pid, "Success:", e.Success)
			}
		}

		if len(e.Errors) > 0 {
			log.Println("ERROR -> Cmd:", e.Cmd, "Args:", e.Args, "Errors:", e.Errors)
		}

		select {
		case executions <- e:
		case <-done:
			return
		}
	}
}

func controller(c Commands, outJSON string) error {

	done := make(chan struct{})
	defer close(done)

	commands, errc := dispatchCommands(done, c)

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(routines)

	executions := make(chan Execution)

	for i := 0; i < routines; i++ {
		go func() {
			commandDigester(done, commands, executions)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(executions)
	}()

	var err error
	var fJ *os.File
	var prep []byte
	var cont int
	var fail int

	bString := []byte("[\n")
	sString := []byte(",\n")
	eString := []byte("\n]\n")

	first := true
	writeJSON := false

	if outJSON != "" {
		fJ, err = os.Create(outJSON)
		if err != nil {
			log.Println(err)
		} else {
			writeJSON = true
			defer func(fJ *os.File) { fJ.Close() }(fJ)
		}
	}

	for e := range executions {
		cont++
		if !e.Success {
			fail++
		}

		if !writeJSON {
			continue
		}

		if first {
			prep = bString
		} else {
			prep = sString
		}

		first = false

		prettyJSON, err := json.MarshalIndent(e, "", "  ")
		if err != nil {
			log.Printf("Can't encode json response for PID: %d\n", e.Pid)
			continue
		}

		_, err = fJ.Write(append(prep, prettyJSON...))
		if err != nil {
			log.Println("Error when writing JSON ", outJSON, ": ", err.Error())
		} else {
			fJ.Sync()
		}
	}

	if !first {
		_, err = fJ.Write(eString)
		if err != nil {
			log.Println("Error when writing JSON ", outJSON, ": ", err.Error())
		}
	}

	if err = <-errc; err != nil {
		log.Println(err)
	}

	totalSeconds := int(time.Since(start).Seconds())

	log.Printf("Final -> Executions: %d Success: %d Fail: %d Elapsed: %d\n", cont, cont-fail, fail, totalSeconds)

	if err != nil || fail > 0 {
		return errors.New("errors in execution/s")
	}
	return nil
}

func main() {
	flag.Parse()
	if inJSON == "" {
		flag.PrintDefaults()
		os.Exit(5)
	}

	if outJSON != "" {
		err := isUsable(outJSON, true)
		if err != nil {
			log.Fatal(err)
		}
	}

	c, err := deserializeJSON(inJSON)
	if err != nil {
		log.Fatal(err)
	}

	err = controller(c, outJSON)
	if err != nil {
		log.Fatal(err)
	}
}
