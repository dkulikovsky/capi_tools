package main

import (
	"./capi"
	"./clusterapi"
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/kr/pretty"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	log.Println("starting holy mess")

	// command line options magic
	// capi: url to capi, default = https://capi-dev.yandex-team.ru:29001
	// instance: instance file with task description and other usefull options like owner and scheduler id

	// parse input config for instance
	// check workload mandatory options and set defaults

	owner := new(clusterapi.Owner)
	owner.OwnerId = "dkulikovsky_task_owner_id"
	owner.Priority = 100
	owner.ProjectId = "CAPIDEVNETS"

	// at this moment we have a parsed workload object
	// now it is just a statically configured workload
	workload := capi.Workload(owner)

	// get cluster state compact view

	// run first fit algorithm on cluster state to get first host that matches resources requested

	// now we have a host and it's etag and now ready to apply transition to cluster state
	var host_etag int64
	host_etag = 0
	host := "myt1-0633-23059.vm.search.yandex.net"

	// update workload params
	workload = capi.SetWlHost(host, workload)
	// get ip and hostname from ip-broker
	workload = capi.SetWlNet("2a02:6b8:c03:22d:0:4097:5324:2bf5", "i-53242bf59a04.qloud-c.yandex.net", workload)

	group := capi.GroupTransition(host, host_etag, workload, owner)
	apply := capi.ApplyGroup(group)
	log.Printf("Got instance object:\n %# v \n", pretty.Formatter(*apply))

	data, err := proto.Marshal(apply)
	if err != nil {
		log.Fatal("marshaling err: %v\n", err)
	}

	// send apply request to capi
	capi_url := "http://sit-dev-01-sas.haze.yandex.net:8082/proto/v0/apply/group"
	resp, err := http.Post(capi_url, "", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Failed with applyGroup request: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	fmt.Println("The calculated length is:", len(string(body)))
	fmt.Println("   ", resp.StatusCode)
	hdr := resp.Header
	for key, value := range hdr {
		fmt.Println("   ", key, ":", value)
	}

	log.Println("successfully marshaled empty container obj")
}
