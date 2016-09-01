package capi

import (
	"capi_tools/clusterapi"
	"math/rand"
    "time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// rune slice for uniq id generator
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GroupTransition(host string, etag int64, wl *clusterapi.Workload, owner *clusterapi.Owner) *clusterapi.GroupTransition {
	trans := new(clusterapi.Transition)
	trans.HostId = host
	trans.HostStateEtag = etag
	trans.Workloads = append(trans.Workloads, wl)

	gtrans := new(clusterapi.GroupTransition)
	gtrans.Owner = owner
	gtrans.GroupId = "dkulikovsky_group_" + gen_id()
	gtrans.GroupOperationId = "dkulikovsky_group_operation_" + gen_id()
	gtrans.Transitions = append(gtrans.Transitions, trans)
	return gtrans
}


func ApplyGroup(group *clusterapi.GroupTransition) *clusterapi.ApplyGroupTransitionRequest {
	return &clusterapi.ApplyGroupTransitionRequest{
		SchedulerSignature: sched_sign(),
		GroupTransitions:   []*clusterapi.GroupTransition{group}}
}


func Workload(owner *clusterapi.Owner) *clusterapi.Workload {
	wl := new(clusterapi.Workload)
	wl.Owner = owner
	wl.TargetState = "ACTIVE"

	wl.Entity = new(clusterapi.Entity)
	wl.Entity.Instance = get_instance()

	wl_id := new(clusterapi.WorkloadId)
	wl_id.Slot = &clusterapi.Slot{Service: "dkulikovsky_test_service_v0.1.0"}
	wl_id.Configuration = &clusterapi.ConfigurationId{GroupId: "dkulikovsky_configuration_" + gen_id()}
	wl.Id = wl_id

	return wl
}


func SetWlHost(host string, wl *clusterapi.Workload) *clusterapi.Workload { 
    wl.Id.Slot.Host = host
    return wl
}


func SetWlNet(ip, hostname string, wl *clusterapi.Workload) *clusterapi.Workload {
    net := map[string]string{
		"meta.net":                        "macvlan vlan1478 eth0",
		"meta.ip":                         "eth0 "+ ip,
		"meta.hostname":                   hostname,
	}
    for k,v := range net {
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

func get_container(constraints map[string]string) *clusterapi.Container {
	c := new(clusterapi.Container)
	// set resources needed by intance (Container)
	resources := &clusterapi.ComputingResources{
		CpuPowerPercentsCore: 50,
		RamBytes:             8589934592,
	}
	c.ComputingResources = resources
	c.Constraints = constraints
	c.Id = "dkulikovsky_container_" + gen_id()

	return c
}

func get_common_root() *clusterapi.Volume {
	// construct common root volume
	v := new(clusterapi.Volume)
	v.MountPoint = "/"
	v.Uuid = "dkulikovsky_volume_" + gen_id()

	// create layers
	var layers []*clusterapi.Resource
	layers = append(layers, &clusterapi.Resource{
		Uuid: "ubuntu-precise",
		Urls: []string{"rbtorrent:a3a80ac6aba30bd8350cfa3f56488bcc4615e0f7"},
	})
	v.Layers = layers

	return v
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

func get_instance() *clusterapi.Instance {
	// create instance
	instance := new(clusterapi.Instance)

	// create container
	constraints := map[string]string{
		// "memory_limit":                    "8589934593",
		"meta.net":                        "macvlan vlan1478 eth0",
		"meta.ip":                         "eth0 2a02:6b8:c03:21d:0:4097:de37:e7e4",
		"meta.hostname":                   "i-de37e7e414b0.qloud-c.yandex.net",
		"meta.virt_mode":                  "os",
		"meta.command":                    "/sbin/init",
	}
	instance.Container = get_container(constraints)

	// chroot
	instance.Volumes = append(instance.Volumes, get_common_root())

	// start hook
	instance.Resources = make(map[string]*clusterapi.Resourcelike)
	instance.Resources["iss_hook_start"] = get_dummy_start_hook()

	return instance
}


func sched_sign() *clusterapi.SchedulerSignature {
	return &clusterapi.SchedulerSignature{SchedulerId: "dkulikovsky_scheduler_id"}
}


