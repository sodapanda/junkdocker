#!/usr/bin/bash
echo "start gost portforward"
DST=$1
echo ${DST}
tc qdisc add dev eth0 root tbf rate 8mbit latency 1ms burst 1048576
./gost -L=tcp://:8800/${DST}