package state

import (
	"../clusterapi"
	"bytes"
	"github.com/golang/protobuf/proto"
	"github.com/kr/pretty"
	"io/ioutil"
	"log"
	"net/http"
)

type Resource struct {
	Cpu      uint32
	Mem      uint64
	Ipv4     bool
	Ipv6     bool
	IoRead   uint32
	IoWrite  uint32
	Net      uint64
	HddSpace uint64
	Tags     map[string]string
}

type Host struct {
	Id        string
	Etag      int64
	Resources map[string]*Resource
	Health    string
	Location  string
}

func GetCompactState() []*Host {
	var raw_state *clusterapi.ClusterState
	var result []*Host

	// prepare clusterstate request to capi
	get_raw_state(raw_state)
	for _, host := range raw_state.Hosts {
		h := new(Host)
		h.Id = host.Metadata.Id
		h.Etag = host.Metadata.Etag
		h.Health = clusterapi.HostHealthState_name[host.Metadata.Health.State]
		h.Resources = decode_resources(host.Metadata.ComputingResources)

		// now update each host resource accordingly to running workloads
		for wl := range host.Workloads {
            var err string
			h.Resources, err = deduct_resources(h.Resources, wl.Entity.Instance.Container.ComputingResources)
			if err != nil {
				log.Printf("Failed to deduct resources with %s and workload %s\n", pretty.Formatter(h.Resources), pretty.Formatter(wl.Entity.Instance.Container.ComputingResources))
			}
		}

	}
	return nil
}

func deduct_resources(res Host.Resources, wl_res *clusterapi.ComputingResources) (Host.Resources, string) {
	return nil, nil
}

func get_raw_state(raw_state *clusterapi.ClusterState) {
	req := new(clusterapi.GetStateRequest)
	data, err := proto.Marshal(req)
	if err != nil {
		log.Fatal("marshaling err: %v\n", err)
	}

	// send state request to capi
	capi_url := "http://sit-dev-01-sas.haze.yandex.net:8082/proto/v0/state/full"
	resp, err := http.Post(capi_url, "", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("failed to issue state request: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to read response request: %v\n", err)
	}

	// debug
	log.Println("The calculated length is:", len(string(body)))
	log.Println("   ", resp.StatusCode)
	hdr := resp.Header
	for key, value := range hdr {
		log.Println("   ", key, ":", value)
	}
	// endof debug

	// unmarshal request
	if err := proto.Unmarshal(body, raw_state); err != nil {
		log.Fatal("failed to unmarshal response: %v\n", err)
	}
}
