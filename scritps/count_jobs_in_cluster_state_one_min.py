#!/usr/bin/env python
import json
import requests
import sys
import re
import socket
import time

def get_cluster(host):
    host_types = { "rtc": re.compile("^.*\.vm\.search\.yandex\.net$"), "r2": re.compile("^s1.*\.qloud\.yandex\.net$"), 
		   "tsnet": re.compile("^tsnet.*search\.yandex\.net$"), "qloud": re.compile("^pool.*\.qloud\.yandex\.net$"),
		   "zerling": re.compile("^zergling.*$") }
    for t in host_types.keys():
        if host_types[t].match(host):
            return t
    return "unknown"


def get_wl_scheduler(wl):
    try:
        res = wl['schedulerId']['name']
    except KeyError:
        return "unknown"
    return res

def get_wl_ram(wl):
    try:
        res = wl['computingRequirements']['resources']['ru.yandex.schedulers.cluster.api.computing.RAM']['capacity']
    except KeyError:
        return 0 
    return res
    
def get_wl_disk(wl):
    try:
        res = wl['computingRequirements']['resources']['ru.yandex.schedulers.cluster.api.computing.HDDSpace']['capacity']
    except KeyError:
        return 0
    return res

def get_wl_cpu(wl):
    try:
        res = wl['computingRequirements']['resources']['ru.yandex.schedulers.cluster.api.computing.CPUPower']['powerPercents']
    except KeyError:
        return 0
    return res

def prepare_result(data, hosts_num):
    prefix = "one_min.capi"
    result = []
    ts = int(time.time())
    for c in data.keys():
        for sc in data[c].keys():
            for k in data[c][sc].keys():
                result.append("%s.%s.%s.%s %s %s" % (prefix, c, sc, k, data[c][sc][k], ts))

    for c in hosts_num.keys():
        for state in hosts_num[c].keys():
            result.append("%s.%s.hosts.%s %s %s" % (prefix, c, state, hosts_num[c][state], ts))

    return result
    
def send_to_graphite(output):
    log_msg("debug","\n".join(output))
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect(('localhost', 2024))
        s.sendall("%s\n\n" % "\n".join(output))
        s.close()
    except:
        return 0

def log_msg(lvl, msg):
    if re.search('debug', lvl, re.I) and not DEBUG:
        return
    print "%s [%s] %s" % (time.ctime(), lvl, msg)


if __name__ == "__main__":
    log_msg("info", "start")   
    DEBUG = 0
    # get current cluster state
    r = requests.get("http://capi-sas.yandex-team.ru:29100/rest/v0/state/0")
    r.encoding = "utf-8"
    r.raise_for_status()
    log_msg("info", "got json from capi")   
    # try to load json
    try:
        data = json.loads(r.text)
    except Exception, e:
        print "Failed to load data, %s" % e
        sys.exit(1)
    
    try:
        f = open("/var/tmp/cluster_state_%s.json" % int(time.time()), "w")
        f.write(json.dumps(data, indent=4))
    except:
        log_msg("warn", "failed to write cluster state json")
    finally:
        f.close()
    
    log_msg("info", "loaded json")   
    # load local json
#    f = "".join(open("cluster_state.json", "r").readlines())
#
#    # try to load json
#    try:
#       data = json.loads(f)
#    except Exception, e:
#       print "Failed to load data, %s" % e
#       sys.exit(1)

    # ok, now we have json load to memory, lets traverse
    # result must be smthn like this: result = { cluster_A: [ scheduler1: { num_of_jobs: N, cpu_alloc: F, mem_alloc: F } , scheduler2: num_of_jobs... ], cluster_B: [ ..] }
    # resources allocated by scheduler: cpu, ram

    result = {}
    hosts_num = {}
    for host in data['hosts'].keys():
        # match host to cluster
        cluster = get_cluster(host)
        if not result.has_key(cluster):
            result[cluster] = {}

        # count number of hosts in each cluster
        host_state = data['hosts'][host]["hostHealth"].get("state", "unknown")
        hosts_num.setdefault(cluster, {})
        if not hosts_num[cluster].has_key(host_state):
            hosts_num[cluster][host_state] = 0
        hosts_num[cluster][host_state] += 1

        # go through workloads list
        for wl in data['hosts'][host]['entities']:
            # get workload params
            scheduler = get_wl_scheduler(wl)
            ram  = get_wl_ram(wl)
            cpu  = get_wl_cpu(wl)
            disk = get_wl_disk(wl)
            
            # set default
            if not result[cluster].has_key(scheduler):
                result[cluster][scheduler] = { "number_of_jobs": 0, "cpu_alloc": 0, "mem_alloc": 0, "disk_alloc": 0 }
            
            # update result
            result[cluster][scheduler]["number_of_jobs"] += 1 
            result[cluster][scheduler]["cpu_alloc"]      += float(cpu)
            result[cluster][scheduler]["mem_alloc"]      += float(ram)
            result[cluster][scheduler]["disk_alloc"]     += int(disk)

    log_msg("info", "parsed to result")   

    send_to_graphite(prepare_result(result, hosts_num))

    try:
        f = open("/var/tmp/result_%s.json" % int(time.time()), "w")
        f.write(json.dumps(result, indent=4))
    except:
        log_msg("warn", "failed to write cluster state json")
    finally:
        f.close()
 
    log_msg("info", "sent result")   
    log_msg("info", "done")   
