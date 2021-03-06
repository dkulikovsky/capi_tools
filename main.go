package main

import (
	"bytes"
	"capi_tools/capi/capi"
//	"capi_tools/capi/sched"
	"capi_tools/capi/state"
	"capi_tools/clusterapi"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/kr/pretty"
	"io/ioutil"
	"log"
	"net/http"
)
var capi_url string = "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0/state/full"

//var capi_url string "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
	owner := &clusterapi.Owner{
		OwnerId:   "dkulikovsky_task_owner_id",
		Priority:  100,
		ProjectId: "CAPIDEVNETS",
	}

	var cstate []*state.Host
	cstate = state.GetCompactState(capi_url)
	PrintState(cstate)
	run_sample_workload(cstate, owner)
}

func run_sample_workload(cstate []*state.Host, owner *clusterapi.Owner) {
	log.Println("starting holy mess")
	workload := capi.SampleWorkload(owner)

    // predefined host, but with real etag from cstate
	host := "myt1-0633-23059.vm.search.yandex.net"
    host_etag, ok := get_host_etag(host)
    if !ok {
        log.Fatal("Failed to find host %s in cstate", host)
    }

	// update workload params
	// get ip and hostname from ip-broker
	workload = capi.SetWlHost(host, workload)
	workload = capi.SetWlNet("2a02:6b8:c03:22d:0:4097:5324:2bf5", "i-53242bf59a04.qloud-c.yandex.net", workload)

	group := capi.GroupTransition(host, host_etag, workload, owner)
	apply := capi.ApplyGroup(group)
	log.Printf("Got instance object:\n %# v \n", pretty.Formatter(*apply))

	data, err := proto.Marshal(apply)
	if err != nil {
		log.Fatal("marshaling err: %v\n", err)
	}

	// send apply request to capi
	resp, err := http.Post(capi_url, "", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Failed with applyGroup request: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to read response request: %v\n", err)
	}

	fmt.Println("The calculated length is:", len(string(body)))
	fmt.Println("   ", resp.StatusCode)
	hdr := resp.Header
	for key, value := range hdr {
		fmt.Println("   ", key, ":", value)
	}
    log.Printf("body %s", body)
    log.Printf("parse response obj %v", parse_apply_res(body))
	log.Println("successfully marshaled empty container obj")
}

type ApplyResponse struct {
    GroupId string
    Exception string
}

func get_host_etag(host string, cstate []*state.Host) uint64, bool {
    for _,h := range cstate {
        if h.Id == host {
            return h.Etag, true
        }
    }
    return 0, false
}

func parse_apply_res(body []byte) []*ApplyResponse {
    result := make([]*ApplyResponse,0)
	// unmarshal request
	raw := new(clusterapi.ApplyGroupTransitionResponse)
	if err := proto.Unmarshal(body, raw); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}
    return result
}
