package main

import (
	"capi/state"
	"log"
)

var capiURL = "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0"

//var capiURL = "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
	// prepare clusterstate request to capi
	filter := state.Filter{
		Host: "all",
		Wl:   "all",
	}

	cstate, err := state.GetCompactState(filter, capiURL)
	if err != nil {
		log.Fatalf("Failed to get state: %v", err)
	}
	state.PrintState(cstate)
}
