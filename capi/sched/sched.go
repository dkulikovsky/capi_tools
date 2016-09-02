package sched

import (
	"capi_tools/capi/state"
	"capi_tools/clusterapi"
	"log"
)

func first_fit(workload *clusterapi.Workload, state []*state.Host) string {
	for _, host := range state {
		if fit_in(workload, host) {
			log.Printf("found matching host for wl: %s\n", host.Id)
			return host.Id
		}
	}
	return ""
}

func fit_in(workload *clusterapi.Workload, target_host *state.Host) bool {
	wl_res := state.DecodeResources(workload.Entity.Instance.Container.ComputingResources)
	if wl_res.Cpu > target_host.ResFree.Cpu {
		return false
	}
	if wl_res.Mem > target_host.ResFree.Mem {
		return false
	}
	if wl_res.IoRead > target_host.ResFree.IoRead {
		return false
	}
	if wl_res.IoWrite > target_host.ResFree.IoWrite {
		return false
	}
	if wl_res.HddSpace > target_host.ResFree.HddSpace {
		return false
	}
	if wl_res.Net > target_host.ResFree.Net {
		return false
	}
	// check bool constraints

	// check tags constraints

	return true
}
