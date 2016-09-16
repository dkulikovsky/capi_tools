#!/usr/bin/env python
import json
import requests
import sys
import socket
import numpy as np
import matplotlib.pyplot as plt
from optparse import OptionParser

if __name__ == "__main__":
    parser = OptionParser()
    parser.add_option("-f", "--file", dest="filename", help="write hist to file as png",)
    parser.add_option("-c", "--capi", dest="capi", default="capi-sas.yandex-team.ru", 
                        help="capi host, default: capi-sas.yandex-team.ru")
    parser.add_option("--state", dest="state_file", help="cluster state json file, for offline stat")
    (options, args) = parser.parse_args()

    # get current cluster state
    if options.state_file:
        try:
            f = "".join(open(options.state_file, "r").readlines())
        except Exception, e:
            print "Failed to read file %s, error: %s" % (options.state_file, e)
            sys.exit(1)
        # try to load json
        try:
            data = json.loads(f)
        except Exception, e:
            print "Failed to load data, %s" % e
            sys.exit(1)
    else:
        url = "http://" + options.capi + ":29100/rest/v0/state/0"
        r = requests.get(url)
        r.encoding = "utf-8"
        r.raise_for_status()
        # try to load json
        try:
            data = json.loads(r.text)
        except Exception, e:
            print "Failed to load data, %s" % e
            sys.exit(1)

    cpu = []
    ram = []
    for host in data['hosts'].keys():
        # count number of hosts in each cluster
        host_ram = data['hosts'][host]["computingResources"]["resources"]["ru.yandex.schedulers.cluster.api.computing.RAM"].get("capacity", 0)
        host_cpu = data['hosts'][host]["computingResources"]["resources"]["ru.yandex.schedulers.cluster.api.computing.CPUPower"].get("powerPercents", 0)
        ram.append(host_ram)
        cpu.append(host_cpu)

    ram_gb = []
    for i in ram:
        ram_gb.append(i/1024/1024/1024)

    yticks = np.arange(0, 4000, 100)

    # plot ram
    plt.subplot(2, 1, 1)
    plt.title("Ram GB by host")
    plt.xlabel("Ram GB")
    plt.ylabel("hosts")
    plt.grid(True)
    xticks = np.arange(0, 300, 10)
    plt.xticks(xticks)
    plt.yticks(yticks)
    plt.hist(ram_gb, bins=xticks)

    # plot cpu
    plt.subplot(2,1,2)
    xticks = np.arange(0, np.max(cpu), 200)
    yticks = np.arange(0, 5000, 200)
    plt.title("CPU % by host")
    plt.xlabel("CPU %")
    plt.ylabel("hosts")
    plt.xticks(xticks)
    plt.yticks(yticks)
    plt.hist(cpu, bins=int(np.max(cpu)/200))
    plt.grid(True)

    if options.filename:
	fig = plt.gcf()
	fig.set_size_inches(20,12)
        plt.savefig(options.filename)
    else:
        plt.show()
