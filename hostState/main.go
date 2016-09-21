package main

import (
	"capi/state"
	"flag"
	"fmt"
	"os"
)

var capiURL = flag.String("capi", "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0", "capi host url")
var host = flag.String("host", "", "host to inspect")

//var capi_url string "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
	flag.Parse()
	if *host == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	filter := state.Filter{
		Host: fmt.Sprintf("'HostMetadata/id' == '%s'", *host),
		Wl:   "all",
	}
	cstate, err := state.GetCompactState(filter, *capiURL)
	if err != nil {
		fmt.Printf("Failed to run task on capi %s, reason: %v", *capiURL, err)
	}
	state.PrintState(cstate)
}
