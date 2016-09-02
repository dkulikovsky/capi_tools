package main

import (
	"bytes"
	"capi_tools/capi/capi"
	"capi_tools/capi/state"
	"capi_tools/clusterapi"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
)

type ApplyResponse struct {
    GroupId string
    Exception string
}

var capi_url string = "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0/state/full"
//var capi_url string "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"

func main() {
	log.Println("starting holy mess")
	owner := &clusterapi.Owner{
		OwnerId:   "dkulikovsky_task_owner_id",
		Priority:  100,
		ProjectId: "CAPIDEVNETS",
	}
    cstate := state.GetCompactState(capi_url)
    state.PrintState(cstate)
	run_sample_workload(cstate, owner)
}

func run_sample_workload(cstate []*state.Host, owner *clusterapi.Owner) {
	workload := capi.SampleWorkload(owner)
    // predefined host, but with real etag from cstate
	host := "s1-1106.qloud.yandex.net"
    host_etag, ok := get_host_etag(host, cstate)
    if !ok {
        log.Printf("Failed to find host %s in cstate", host)
    }
    log.Printf("got host %s etag %d", host, host_etag)

	// update workload params
	// get ip and hostname from ip-broker
	workload = capi.SetWlHost(host, workload)
	workload = capi.SetWlNet("2a02:6b8:c02:1:0:4097:b0ce:c94d", "i-b0cec94d0036.qloud-c.yandex.net", workload)

	group := capi.GroupTransition(host, host_etag, workload, owner)
	apply := capi.ApplyGroup(group)
	data, err := proto.Marshal(apply)
	if err != nil {
		log.Fatal("marshaling err: %v\n", err)
	}

	// send apply request to capi
	client := &http.Client{}
    req, err := http.NewRequest("POST", capi_url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("failed to create new request: ", err)
	}
	req.Header.Add("Accept", "application/x-protobuf")
	req.Header.Add("Content-Type", "application/x-protobuf")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("failed to issue state request: %v\n", err)
	}
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

func get_host_etag(host string, cstate []*state.Host) (int64, bool) {
    for _,h := range cstate {
        if h.Health == "UP" && h.Id == host {
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
