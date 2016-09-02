package state

import (
	"bytes"
	"capi_tools/clusterapi"
	"fmt"
	"github.com/golang/protobuf/proto"
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
	ResTotal  *Resource
	ResFree   *Resource
	Health    string
	Location  string
	Workloads []*clusterapi.Workload
}

func GetCompactState() []*Host {
	var raw_state *clusterapi.ClusterState
	var result []*Host

	// prepare clusterstate request to capi
	raw_state = get_raw_state()
	for _, host := range raw_state.Hosts {
		h := new(Host)
		h.Id = host.Metadata.Id
		h.Etag = host.Metadata.Etag
		h.Health = clusterapi.HostHealthState_name[int32(host.Metadata.Health.State)]
		h.ResTotal = DecodeResources(host.Metadata.ComputingResources)
		h.ResFree = new(Resource)
		*h.ResFree = *h.ResTotal
		h.Workloads = host.Workloads

		// now update each host resource accordingly to running workloads
		for _, wl := range host.Workloads {
			var wl_res *Resource
			if wl.Entity.Instance != nil {
				wl_res = DecodeResources(wl.Entity.Instance.Container.ComputingResources)
			} else {
				wl_res = DecodeResources(wl.Entity.Job.Container.ComputingResources)
			}
			fmt.Printf("workload resources: host: %s, cpu: %d, mem: %d\n", h.Id, wl_res.Cpu, wl_res.Mem)

			h.ResFree = deduct_resources(h.ResFree, wl_res)
		}
		result = append(result, h)
	}
	return result
}

func DecodeResources(res *clusterapi.ComputingResources) *Resource {
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

func get_raw_state() *clusterapi.ClusterState {
	get_state_req := new(clusterapi.GetStateRequest)
	get_state_req.HostFilter = "all"
	get_state_req.WorkloadFilter = "all"
	get_state_req.PreviousVersion = new(clusterapi.ClusterVersion)
	get_state_req.PreviousVersion.Versions = make(map[string]uint64)
	get_state_req.PreviousVersion.Versions["version"] = 0
	data, err := proto.Marshal(get_state_req)
	if err != nil {
		log.Fatal("marshaling err: %v\n", err)
	}

	// send state request to capi
	capi_url := "http://sit-dev-01-sas.haze.yandex.net:8081/proto/v0/state/full"

	//	capi_url := "http://iss00-prestable.search.yandex.net:8082/proto/v0/state/full"
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

	// debug
	log.Println("The calculated length is:", len(string(body)))
	log.Println("   ", resp.StatusCode)
	hdr := resp.Header
	for key, value := range hdr {
		log.Println("   ", key, ":", value)
	}
	// endof debug

	// unmarshal request
	raw_state := clusterapi.ClusterState{}
	if err := proto.Unmarshal(body, &raw_state); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}
	return &raw_state
}
