package main

import (
	"capi/state"
	"capi/task"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kr/pretty"
	yaml "gopkg.in/yaml.v2"
)

var capiURL = flag.String("capi", "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0", "capi host url")
var taskF = flag.String("task", "", "path to task.yaml")

//var capi_url string "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
	flag.Parse()
	if *taskF == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// parse yaml and create Task
	task := &task.Task{}

	// TODO: fix to proper args reader
	source, err := ioutil.ReadFile(*taskF)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(source, task)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// debug
	log.Printf("--- config:\n%# v\n\n", pretty.Formatter(task))

	result, err := state.TaskInfo(task, *capiURL)
	if err != nil {
		fmt.Printf("Failed to run task on %s, reason: %v", *capiURL, err)
	}
	fmt.Printf("Got feedbacks:\n=====Feedback on task=====\n%# v\n=====END=====\n", pretty.Formatter(result))
}
