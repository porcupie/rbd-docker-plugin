package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	defaultShellTimeout = 2 * 60 * time.Second
)

// sh is a simple os.exec Command tool, returns trimmed string output
func sh(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	log.Printf("INFO: sh CMD: %q", cmd)
	out, err := cmd.Output()
	log.Printf("INFO: [out, err]/[%s, %s]", out, err)
	return strings.Trim(string(out), " \n"), err
}

// ShResult used for channel in timeout
type ShResult struct {
	Output string // STDOUT
	Err    error  // go error, not STDERR
}

type ShTimeoutError struct {
	timeout time.Duration
}

func (e ShTimeoutError) Error() string {
	return fmt.Sprintf("Reached TIMEOUT on shell command")
}

// shWithDefaultTimeout will use the defaultShellTimeout so you dont have to pass one
func shWithDefaultTimeout(name string, args ...string) (string, error) {
	return shWithTimeout(defaultShellTimeout, name, args...)
}

// shWithTimeout will run the Cmd and wait for the specified duration
func shWithTimeout(howLong time.Duration, name string, args ...string) (string, error) {
	// duration can't be zero
	if howLong <= 0 {
		return "", fmt.Errorf("Timeout duration needs to be positive")
	}
	// set up the results channel
	resultsChan := make(chan ShResult, 1)
	if isDebugEnabled() {
		log.Printf("DEBUG: shWithTimeout: %v, %s, %v", howLong, name, args)
	}

	// fire up the goroutine for the actual shell command
	go func() {
		out, err := sh(name, args...)
		resultsChan <- ShResult{Output: out, Err: err}
	}()

	select {
	case res := <-resultsChan:
		return res.Output, res.Err
	case <-time.After(howLong):
		return "", ShTimeoutError{timeout: howLong}
	}

	return "", nil
}

// grepLines pulls out lines that match a string (no regex ... yet)
func grepLines(data string, like string) []string {
	var result = []string{}
	if like == "" {
		log.Printf("ERROR: unable to look for empty pattern")
		return result
	}
	like_bytes := []byte(like)

	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		if bytes.Contains(scanner.Bytes(), like_bytes) {
			result = append(result, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("WARN: error scanning string for %s: %s", like, err)
	}

	return result
}

// regexpLines pulls out lines that match a regexp as group matches
func regexpLines(data string, regexp_s string) [][]string {
	var result = [][]string{}

	r, err := regexp.Compile(regexp_s)
	if err != nil {
		log.Printf("ERROR: unable to compile regexp")
		return result
	}

	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		s := scanner.Text()
		if r.MatchString(s) {
			result = append(result, r.FindStringSubmatch(s))
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("WARN: error scanning string for %s: %s", regexp_s, err)
	}

	return result
}

// Linux process management
//ref: https://github.com/kimpettersen/GoProcs/blob/master/src/procs.go
type Process struct {
	Pid        string
	Executable string
}

// List all processes in system
func listProcesses() []Process {
	var processes []Process
	files, err := ioutil.ReadDir("/proc")

	if err != nil {
		log.Fatal("Could not read dir /proc :)")
	}
	var proc Process

	for _, file := range files {
		if _, err := strconv.Atoi(file.Name()); err == nil {

			cmd, err := ioutil.ReadFile("/proc/" + file.Name() + "/cmdline")

			cmdString := strings.Join(strings.Split(string(cmd), "\x00"), " ")

			if err != nil {
				log.Fatal("Can't read file:", err)
			}

			proc = Process{
				Pid:        file.Name(),
				Executable: cmdString,
			}
			processes = append(processes, proc)
		}
	}
	return processes
}

// kill a process
func kill(proc Process, signal string) error {
	if err := exec.Command("kill", "-"+signal, string(proc.Pid)).Start(); err != nil {
		log.Fatal("Couldn't kill ", err)
		return err
	}
	fmt.Printf("Killed %v\n", proc.Executable)
	return nil
}
