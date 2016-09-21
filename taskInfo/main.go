package main

import (
	"capi/state"
	"capi/task"
	"io/ioutil"
	"log"
	"os"

	"github.com/kr/pretty"
	yaml "gopkg.in/yaml.v2"
)

var capiURL = "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0"

//var capi_url string "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
	// parse yaml and create Task
	task := &task.Task{}

	// TODO: fix to proper args reader
	filename := os.Args[1]
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(source, task)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// debug
	log.Printf("--- config:\n%# v\n\n", pretty.Formatter(task))

	// run task

	resp, err := state.TaskInfo(task, capiURL)
	if err != nil {
		log.Fatalf("Failed to get task info on capi %s, reason: %v", capiURL, err)
	} else {
		log.Printf("Got task:\n%# v", pretty.Formatter(resp))
	}
}
