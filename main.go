package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

const fileName = "gitconfig.json"
const gitConfigFolder = "/.git/config"

type Config struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	content := buildText(open())
	configFile := getFolderName(os.Args[1])
	err := run(5, "git", "clone", os.Args[1])
	if err != nil {
		log.Println(err)
		return
	}
	err = writeConfig(content, configFile)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Done!")
}

func writeConfig(content []byte, file string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.Write(content); err != nil {
		return err
	}
	return nil
}

func open() Config {
	var c = Config{}
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println(err)
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

func getFolderName(arg string) string {
	result := strings.Split(arg, "/")
	result = strings.Split(result[1], ".")

	return result[0] + gitConfigFolder
}

func run(timeout int, command string, args ...string) error {

	// instantiate new command
	cmd := exec.Command(command, args...)

	// get pipe to standard output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// start process via command
	if err := cmd.Start(); err != nil {
		return err
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
			return err
		}
		return errors.New("timeout reached, process killed")
	case err := <-done:
		if err != nil {
			close(done)
			return err
		}
		return nil
	}
}
