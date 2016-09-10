package sched

import (
	"capi_tools/capi/state"
	"capi_tools/clusterapi"
	"log"
)

var retry int = 3

func Run(task *task.Task, capi_url string) {
	// don't reschedule on host chosen before
	seen := make([]string, 0, 3)

	for i := 0; i < retry; i++ {
		// get compact cluster state
		cstate := state.GetCompactState(capi_url)

		// get first fit host
		host, err := sched.first_fit(task, cstate)
		if err != nil {
			log.Printf("Failed to get host in first fit, reason: %v", err)
			continue
		}

		if seen_host(host, seen) == true {
			continue
		}

		// get ip for task on host
		err = sched.get_net(task, host)
		if err != nil {
			// we failed to get ip for host, may be we will get more luck with other host
			log.Printf("Failed to get network for task on host %s, reason %v", err)
			// most likely ip-broker can't get an ip for host
			// make sure not to use this host any more
			seen = append(seen, host)
			continue // go to the next iteration of scheduling
		}

		// and create grouptransition with Task
		t := capi.GetTransition(task, host)

		// run capi.Apply
		err = capi.Apply(t, capi_url)
		if err != nil {
			log.Printf("Apply failed: %v", err)
			continue
		}
	}
}

func get_net(task *task.Task, host *state.Host) error {
	task.Ip = "2a02:6b8:c02:1:0:4097:b0ce:c94d"
	task.Hostname = "i-b0cec94d0036.qloud-c.yandex.net"
	return nil
}

func first_fit(task *task.Task, state []*state.Host) (*state.Host, error) {
	for _, host := range state {
		if h.Health != "UP" {
			continue
		}
		if fit_in(task, host) {
			log.Printf("found matching host for wl: %s\n", host.Id)
			return host, nil
		}
	}
	return nil, fmt.Errorf("no matching hosts")
}

func fit_in(task *task.Task, target_host *state.Host) bool {
	if task.Resources.Cpu > target_host.ResFree.Cpu {
		return false
	}
	if task.Resources.Mem > target_host.ResFree.Mem {
		return false
	}
	if task.Resources.Disk > target_host.ResFree.Disk {
		return false
	}
	if task.Resources.Net > target_host.ResFree.Net {
		return false
	}
	return true
}

func seen_host(host string, seen []string) bool {
	for _, seen_host := range seen {
		if host == seen_host {
			return true
		}
	}
	return false
}
