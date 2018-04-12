package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

const fileName = "example.json"

func main() {
	content := buildText(open())
	fmt.Println(string(content))
	configFile := getFolderName(os.Args[1])
	fmt.Println(configFile)
	//s := run(5, "git", "clone", os.Args[1])
	//fmt.Printf(s)

}

type Config struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func open() Config {
	var c = Config{}
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return c
	}
	json.Unmarshal(dat, &c)
	return c
}

var tab = []byte{9}
var newLine = []byte{10}

func buildText(c Config) []byte {
	var lines = []string{
		"[user]",
		"name = %s",
		"email = %s",
		"username = %s",
	}

	var content = []byte{}
	for _, line := range lines {
		var value string
		if strings.Contains(line, "=") {
			tag := strings.Split(line, " = ")
			value = getValue(c, tag[0])
			value = fmt.Sprintf(line, value)
			value = string(append(tab, []byte(value)...))
		} else {
			value = line
		}

		content = append(append(content, []byte(value)...), newLine...)
	}

	return content
}

func getValue(c Config, tag string) string {
	rv := reflect.ValueOf(c)
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		if tagV, ok := rt.Field(i).Tag.Lookup("json"); ok {
			if tagV == tag {
				return rv.Field(i).String()
			}
		}
	}

	return ""
}

const gitConfigFolder = "/.git/config"

func getFolderName(arg string) string {
	result := strings.Split(arg, "/")
	result = strings.Split(result[1], ".")

	return result[0] + gitConfigFolder
}

func run(timeout int, command string, args ...string) string {

	// instantiate new command
	cmd := exec.Command(command, args...)

	// get pipe to standard output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "cmd.StdoutPipe() error: " + err.Error()
	}

	// start process via command
	if err := cmd.Start(); err != nil {
		return "cmd.Start() error: " + err.Error()
	}

	// setup a buffer to capture standard output
	var buf bytes.Buffer

	// create a channel to capture any errors from wait
	done := make(chan error)
	go func() {
		if _, err := buf.ReadFrom(stdout); err != nil {
			panic("buf.Read(stdout) error: " + err.Error())
		}
		done <- cmd.Wait()
	}()

	// block on select, and switch based on actions received
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			return "failed to kill: " + err.Error()
		}
		return "timeout reached, process killed"
	case err := <-done:
		if err != nil {
			close(done)
			return "process done, with error: " + err.Error()
		}
		return "process completed: " + buf.String()
	}
	return ""
}
