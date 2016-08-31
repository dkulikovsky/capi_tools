package main

import (
	"./capi"
	"./clusterapi"
	"github.com/golang/protobuf/proto"
	"github.com/kr/pretty"
	"log"
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
	host := "s1-1010.qloud.yandex.net"

    // update workload params
    workload = capi.SetWlHost(host, workload)
    // get ip and hostname from ip-broker
    workload = capi.SetWlNet("2a02:6b8:c03:21d:0:4097:de37:e7e4", "i-de37e7e414b0.qloud-c.yandex.net", workload)

	group := capi.GroupTransition(host, host_etag, workload, owner)
	apply := capi.ApplyGroup(group)

	log.Printf("Got instance object:\n %# v \n", pretty.Formatter(*apply))
	_, err := proto.Marshal(apply)
	if err != nil {
		log.Fatal("marshaling err: %v\n", err)
	}

	log.Println("successfully marshaled empty container obj")
}
