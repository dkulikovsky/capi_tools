package state

import (
	"../../clusterapi"
	"bytes"
	"github.com/golang/protobuf/proto"
	//"github.com/kr/pretty"
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
	HasSsd   bool
	Net      uint64
	HddSpace uint64
	Tags     map[string]uint64
}

type Host struct {
	Id        string
	Etag      int64
	Resources *Resource
	Health    string
	Location  string
    Workloads []*clusterapi.Workload
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
		h.Health = clusterapi.HostHealthState_name[int32(host.Metadata.Health.State)]
		h.Resources = decode_resources(host.Metadata.ComputingResources)
        h.Workloads = host.Workloads

		// now update each host resource accordingly to running workloads
		for _, wl := range host.Workloads {
			wl_res := decode_resources(wl.Entity.Instance.Container.ComputingResources)
			h.Resources = deduct_resources(h.Resources, wl_res)
		}
		result = append(result, h)
	}
	return result
}

func decode_resources(res *clusterapi.ComputingResources) *Resource {
	r := new(Resource)
	// get common resources
	r.Cpu = res.CpuPowerPercentsCore
	r.Mem = res.RamBytes
	r.HddSpace = res.HddSpaceBytes
	r.Net = res.NetworkOutgoingBps

	// disk iops
	r.IoRead = res.IopsRead
	r.IoWrite = res.IopsWrite
	r.HasSsd = res.HasSsd

	// network options
	r.Ipv4 = res.HasIpv4
	r.Ipv6 = res.HasIpv6

	// conductor tags, cms tags, etc.
	r.Tags = get_resource_tags(res)
	return r
}

func get_resource_tags(res *clusterapi.ComputingResources) map[string]uint64 {
	tags := make(map[string]uint64)
	for _, tag := range res.NamedCountables {
		tags[tag.Name] = tag.Capacity
	}
	return tags
}

func deduct_resources(res, wl_res *Resource) *Resource {
	res.Cpu -= wl_res.Cpu
	res.Mem -= wl_res.Mem
	res.Net -= wl_res.Net
	res.HddSpace -= wl_res.HddSpace
	res.IoRead -= wl_res.IoRead
	res.IoWrite -= wl_res.IoWrite
	return res
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
