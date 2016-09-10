package capi

import (
	"capi_tools/capi/task"
	"testing"
)

func TestGetContainer(t *testing.T) {
	// create sample task
	task := task.Task{
		Owner:    "dkulikovsky",
		Command:  "/bin/bash",
		Hostname: "test-host-1",
		Ip:       "1.1.1.1",
		Resources: task.Resources{
			Cpu:  100,
			Ram:  100,
			Net:  100,
			Disk: 100,
		},
	}

	c := get_container(&task)
	if c == nil {
		t.Errorf("got nil result from function")
	}
	// check fields
	if c.Constraints == nil {
		t.Errorf("constraints are not defined")
	}
	if c.Id == "" {
		t.Errorf("id are not defined")
	}
	if c.Constraints["meta.command"] != task.Command {
		t.Errorf("Container.Constraints['meta.command'] != task.Command, got %s, want %s", c.Constraints["meta.command"], task.Command)
	}
	if c.Constraints["meta.hostname"] != task.Hostname {
		t.Errorf("Container.Constraints['meta.hostname'] != task.Hostname, got %s, want %s", c.Constraints["meta.hostname"], task.Hostname)
	}
	want := "eth0 " + task.Ip
	if c.Constraints["meta.ip"] != want {
		t.Errorf("Container.Constraints['meta.ip'] != task.Ip, got %s, want %s", c.Constraints["meta.ip"], want)
	}
	want = string(task.Resources.Ram)
	if c.Constraints["meta.memory_limit"] != want {
		t.Errorf("Container.Constraints['meta.memory_limit'] != task.Resources.Ram, got %v, want %s", c.Constraints["meta.memory_limit"], want)
	}

	// check resources
	if c.ComputingResources == nil {
		t.Error("computing resources are not defined")
	}
	if c.ComputingResources.CpuPowerPercentsCore != task.Resources.Cpu {
		t.Errorf("Container.ComputingResources.CpuPowerPercentsCore != task.Resources.Cpu, got %v, want %v", c.ComputingResources.CpuPowerPercentsCore, task.Resources.Cpu)
	}
	if c.ComputingResources.RamBytes != task.Resources.Ram {
		t.Errorf("Container.ComputingResources.RamBytes != task.Resources.Ram, got %v, want %v", c.ComputingResources.RamBytes, task.Resources.Ram)
	}
}

func TestCreateVolumes(t *testing.T) {
	// sample task with volumes
	task := task.Task{
		Volumes: map[string]task.Volume{
			"ubuntu": task.Volume{Mount: "/", Url: "http://best/url/in/all/worlds"},
		},
	}

	v := create_volumes(&task)
	if v == nil {
		t.Errorf("return value is nil")
	}

	if len(v) == 0 {
		t.Errorf("no volumes defined")
	}
	if len(v) > 1 {
		t.Errorf("bogus output, only one volume must be defined, %v", v)
	}

	// our sample volume is the first and the only one volume defined
	v0 := v[0]
	if v0.Layers == nil {
		t.Errorf("no layers defined")
	}
	if len(v0.Layers) == 0 {
		t.Errorf("empty layers slice")
	}
	if len(v0.Layers) > 1 {
		t.Errorf("bogus layers slice, %v", v)
	}

	// and again our sample volume has only one layer
	l0 := v[0].Layers[0]
	if l0 == nil {
		t.Errorf("got nil layer")
	}
	if l0.Uuid != "ubuntu" {
		t.Errorf("l0.Uuid != 'ubuntu', got %s", l0.Uuid)
	}
}

func TestGetStartHook(t *testing.T) {
    url := "https://paste.yandex-team.ru/147541/text"

    h := get_start_hook(url)
    if h == nil {
        t.Errorf("return value is nil")
    }

    if h.Resource.Uuid != "start_hook_id" {
        t.Errorf("bogus start hook name, got %s, want 'start_hook_id'", h.Resource.Uuid)
    }
    if len(h.Resource.Urls) == 0 || len(h.Resource.Urls) > 1 {
        t.Errorf("bogus urls array, got %v", h.Resource.Urls)
    }
    if h.Resource.Urls[0] != url {
        t.Errorf("wrong url, got %s, want %s", h.Resource.Urls[0], url)
    }
}
