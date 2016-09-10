package main

import (
    "capi_tools/capi/task"
	"io/ioutil"
	"log"
)



var capi_url string = "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0"

//var capi_url string "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
    // parse yaml and create Task
	var task *task.Task

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
    log.Printf("--- config:\n%# v\n\n", pretty.Formatter(config))

    // run task
    err = sched.Run(task, capi_url)
    if err != nil {
        log.Fatalf("Failed to run task on capi %s, reason: %v", capi_url, err)
    }

}
