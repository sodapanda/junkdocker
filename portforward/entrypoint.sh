#!/usr/bin/bash
echo "start gost portforward"
DST=$1
ID=$2
CID="${ID}_${DST}"
touch cid
printf ${CID} >> cid
tc qdisc add dev eth0 root tbf rate 8mbit latency 1ms burst 1048576
nohup ./gost -L=tcp://:8800/${DST} &
./portmonitor