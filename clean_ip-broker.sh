#!/bin/bash
oauth="8980ba4f3b5b4d7a827fbebd9acc8016"
projectid="CAPIDEVNETS"

# get current endpoints and substract ids
curl -s -i -H"Authorization: OAuth $oauth" -XGET  "ip-broker.qloud.yandex.net/network/endpoint/" > /tmp/${projectid}_endpoints.txt
grep -B1 ${projectid} /tmp/${projectid}_endpoints.txt | grep id | sed -e 's/ *"id": "\(.*\)",$/\1/' | uniq > /tmp/${projectid}_endpoint_ids.txt 

# delete them
for i in `cat /tmp/${projectid}_endpoint_ids.txt`; do 
    curl -s -i -H"Authorization: OAuth ${oauth}" -XDELETE  "ip-broker.qloud.yandex.net/network/endpoint/${i}"
done

n=`cat /tmp/${projectid}_endpoint_ids.txt | wc -l`
echo "${n} endpoints for project ${projectid} was deleted from ip-broker"
