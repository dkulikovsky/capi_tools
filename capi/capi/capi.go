package capi

import (
	"bytes"
	"capi_tools/capi/state"
	"capi_tools/capi/task"
	"capi_tools/clusterapi"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// rune slice for uniq id generator
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type ApplyResponse struct {
	GroupId   string
	Exception string
}

func Apply(group *clusterapi.GroupTransition, capi_url string) error {
	apply := &clusterapi.ApplyGroupTransitionRequest{
		SchedulerSignature: sched_sign(),
		GroupTransitions:   []*clusterapi.GroupTransition{group},
	}
	data, err := proto.Marshal(apply)
	if err != nil {
		return err
	}

	// send apply request to capi
	client := &http.Client{}
	req, err := http.NewRequest("POST", capi_url+"/apply/group", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/x-protobuf")
	req.Header.Add("Content-Type", "application/x-protobuf")

	// now do http request
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// everything except 200 is baaad
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to apply transition, status: %s, body: %s", resp.Status, body)
	}

	return nil
}

func GetTransition(task *task.Task, host *state.Host) *clusterapi.GroupTransition {
	owner := &clusterapi.Owner{
		OwnerId:   task.Owner,
		Priority:  100,
		ProjectId: task.ProjectId,
	}
	workload := create_workload(task, owner)
	return create_transition(host, workload, owner)
}

/* end of exported functions */
func create_transition(host *state.Host, wl *clusterapi.Workload, owner *clusterapi.Owner) *clusterapi.GroupTransition {
	trans := &clusterapi.Transition{
		HostId:        host.Id,
		HostStateEtag: host.Etag,
		Workloads:     []*clusterapi.Workload{wl},
	}

	gtrans := &clusterapi.GroupTransition{
		Owner:            owner,
		GroupId:          owner.OwnerId + "_group_" + gen_id(),
		GroupOperationId: owner.OwnerId + "_group_operation_" + gen_id(),
		Transitions:      []*clusterapi.Transition{trans},
	}

	return gtrans
}

func create_workload(task *task.Task, owner *clusterapi.Owner) *clusterapi.Workload {
	wl := &clusterapi.Workload{
		Owner:       owner,
		TargetState: "ACTIVE",
		Entity:      &clusterapi.Entity{Instance: get_instance(task)},
	}

	wl_id := new(clusterapi.WorkloadId)
	wl_id.Slot = &clusterapi.Slot{
		Service: fmt.Sprintf("%s_%s_%s", task.Owner, task.Service, task.Version),
		Host:    task.Hostname,
	}
	wl_id.Configuration = &clusterapi.ConfigurationId{GroupId: task.Owner + "_configuration_" + gen_id()}
	wl.Id = wl_id

	return wl
}

func get_instance(task *task.Task) *clusterapi.Instance {
	// create instance
	instance := new(clusterapi.Instance)

	instance.Volumes = create_volumes(task)

	// add resources: mandatory is iss_hook_start, all other are optional
	instance.Resources = make(map[string]*clusterapi.Resourcelike)
	instance.Resources["iss_hook_start"] = get_start_hook(task.StartHook)

	// create actual container

	instance.Container = get_container(task)

	return instance
}

func get_container(task *task.Task) *clusterapi.Container {
	constraints := map[string]string{
		"meta.memory_limit": string(task.Resources.Ram),
		"meta.net":          "macvlan vlan1478 eth0",
		"meta.virt_mode":    "os",
		"meta.command":      task.Command,
		"meta.ip":           "eth0 " + task.Ip,
		"meta.hostname":     task.Hostname,
	}
	c := &clusterapi.Container{
		Constraints: constraints,
		Id:          task.Owner + "_" + gen_id(),
	}

	// set resources needed by intance (Container)
	resources := &clusterapi.ComputingResources{
		CpuPowerPercentsCore: task.Resources.Cpu,
		RamBytes:             task.Resources.Ram,
	}
	c.ComputingResources = resources

	return c
}

func create_volumes(task *task.Task) []*clusterapi.Volume {
	volumes := make([]*clusterapi.Volume, 0, 1)

	for name, volume := range task.Volumes {
		v := &clusterapi.Volume{
			MountPoint: volume.Mount,
			Uuid:       name,
		}
		// create layers
		layers := []*clusterapi.Resource{&clusterapi.Resource{
			Uuid: name,
			Urls: []string{volume.Url},
		}}
		v.Layers = layers
		volumes = append(volumes, v)
	}
	return volumes
}

func get_start_hook(url string) *clusterapi.Resourcelike {
	r := new(clusterapi.Resourcelike)
	r.Resource = &clusterapi.Resource{
		Uuid: "start_hook_id",
		Urls: []string{url},
	}
	return r
}

func sched_sign() *clusterapi.SchedulerSignature {
	return &clusterapi.SchedulerSignature{SchedulerId: "dkulikovsky_scheduler_id"}
}

func parse_apply_res(body []byte) []*ApplyResponse {
	result := make([]*ApplyResponse, 0)
	// unmarshal request
	raw := new(clusterapi.ApplyGroupTransitionResponse)
	if err := proto.Unmarshal(body, raw); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}
	return result
}

func gen_id() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
