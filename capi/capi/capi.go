package capi

import (
	"capi_tools/clusterapi"
    "capi_tools/task"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// rune slice for uniq id generator
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GroupTransition(host string, etag int64, wl *clusterapi.Workload, owner *clusterapi.Owner) *clusterapi.GroupTransition {
	trans := &clusterapi.Transition{
		HostId:        host,
		HostStateEtag: etag,
		Workloads:     []*clusterapi.Workload{wl},
	}

	gtrans := &clusterapi.GroupTransition{
		Owner:            owner,
		GroupId:          "dkulikovsky_group_" + gen_id(),
		GroupOperationId: "dkulikovsky_group_operation_" + gen_id(),
		Transitions:      []*clusterapi.Transition{trans},
	}

	return gtrans
}

func ApplyGroup(group *clusterapi.GroupTransition) *clusterapi.ApplyGroupTransitionRequest {
	return &clusterapi.ApplyGroupTransitionRequest{
		SchedulerSignature: sched_sign(),
		GroupTransitions:   []*clusterapi.GroupTransition{group}}
}

func SampleWorkload(task *task.Task) *clusterapi.Workload {
	wl := &clusterapi.Workload{
		Owner:       task.Owner,
		TargetState: "ACTIVE",
		Entity:      &clusterapi.Entity{Instance: get_instance(task)},
	}

	wl_id := new(clusterapi.WorkloadId)
	wl_id.Slot = &clusterapi.Slot{Service: fmt.Sprintf("%s_%s_%s", task.Owner, task.Service, task.Version)}
	wl_id.Configuration = &clusterapi.ConfigurationId{GroupId: task.Owner+"_configuration_" + gen_id()}
	wl.Id = wl_id

	return wl
}

func SetWlHost(task *task.Task, wl *clusterapi.Workload) *clusterapi.Workload {
	wl.Id.Slot.Host = task.Hostname
	return wl
}

func SetWlNet(task *task.Task, wl *clusterapi.Workload) *clusterapi.Workload {
	net := map[string]string{
		"meta.ip":       "eth0 " + task.Ip,
		"meta.hostname": task.Hostname,
	}
	for k, v := range net {
		wl.Entity.Instance.Container.Constraints[k] = v
	}
	return wl
}

/* end of exported functions */

func gen_id() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func get_container(task task.Task) *clusterapi.Container {
	constraints := map[string]string{
		"memory_limit":                   string(task.Resources.Ram),
		"meta.net":       "macvlan vlan1478 eth0",
		"meta.virt_mode": "os",
		"meta.command":   task.Spec.Command,
	}
	c := &clusterapi.Container{
		Constraints: constraints,
		Id:          "dkulikovsky_container_" + gen_id(),
	}

	// set resources needed by intance (Container)
	resources := &clusterapi.ComputingResources{
		CpuPowerPercentsCore: task.Resources.Cpu,
		RamBytes:             task.Resources.Ram,
	}
	c.ComputingResources = resources

	return c
}

func get_common_root() *clusterapi.Volume {
	// construct common root volume
	v := &clusterapi.Volume{
		MountPoint: "/",
		Uuid:       "dkulikovsky_volume_" + gen_id(),
	}

	// create layers
	layers := []*clusterapi.Resource{&clusterapi.Resource{
		Uuid: "ubuntu-precise",
		Urls: []string{"rbtorrent:a3a80ac6aba30bd8350cfa3f56488bcc4615e0f7"},
	}}
	v.Layers = layers

	return v
}

func get_instance() *clusterapi.Instance {
	// create instance
	instance := new(clusterapi.Instance)

	instance.Volumes = []*clusterapi.Volume{get_common_root()}

	// add resources: mandatory is iss_hook_start, all other are optional
	instance.Resources = make(map[string]*clusterapi.Resourcelike)
	instance.Resources["iss_hook_start"] = get_dummy_start_hook()

	// create actual container

	instance.Container = get_container(constraints)

	return instance
}

func get_dummy_start_hook() *clusterapi.Resourcelike {
	// there must be a iss_hook_start resource to make instance launchable
	r := new(clusterapi.Resourcelike)
	r.Resource = &clusterapi.Resource{
		Uuid: "start_hook_id",
		Urls: []string{"https://paste.yandex-team.ru/147541/text"},
	}
	return r
}

func sched_sign() *clusterapi.SchedulerSignature {
	return &clusterapi.SchedulerSignature{SchedulerId: "dkulikovsky_scheduler_id"}
}
