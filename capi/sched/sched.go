package sched

import ( 
    "capi_tools/clusterapi"
    "capi_tools/capi/state"
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
    if wl_res.Mem > target_host.ResFree.Mem{
        return false
    }
    return true
   }


