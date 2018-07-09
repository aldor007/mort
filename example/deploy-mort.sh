#!/bin/bash
docker pull aldor007/mort:{{ mort_version }}
docker build . -t mort
docker ps | grep mort | awk '{print $1 }' | xargs docker stop
docker ps -a | grep mort | awk '{print $1 }' | xargs docker rm
rm -rf /var/run/mort/mort.sock
docker run  --name mort -d  -p 127.0.0.1:8080  -p 127.0.0.1:8081 -v /var/mort/data/:/data/buckets -v /var/run/mort/:/var/run/mort mort
counter=0
while [ ! -f /var/run/mort/mort.sock ]; do
    sleep 1
    counter=$counter+1
    if (( $counter > 10 )); then
        setfacl -m m::rwx /var/run/mort/mort.sock
        exit $?
    fi
    sleep 1
done
setfacl -m m::rwx /var/run/mort/mort.sock