package main

import (
	"capi_tools/capi/state"
)
var capi_url string = "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0/state/full"
//var capi_url string "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
    state.PrintState(state.GetCompactState(capi_url))
}
