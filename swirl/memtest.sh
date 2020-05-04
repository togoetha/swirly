#!/bin/bash

if [[ -z "${5:-}" ]]; then
  echo "Use: memtest.sh minenodes maxenodes minfnodes maxfnodes iterations"
  exit 1
fi

minenodes=${1}
shift
maxenodes=${1}
shift
minfnodes=${1}
shift
maxfnodes=${1}
shift
iterations=${1}

#pids=$(echo $csvpids | tr ";" "")
#IFS="," read -ra pids <<< "$csvpids"

enodes=$minenodes

while [ $enodes -le $maxenodes ]
do
  fnodes=$minfnodes
  while [ $fnodes -le $maxfnodes ]
  do
    iter=0
    echo "{ \"minEdgeNodes\": $enodes,\"maxEdgeNodes\": 200000,\"edgeNodeStep\": 50000,\"minFogNodes\": $fnodes,\"maxFogNodes\": 400,\"fogNodeStep\": 50,\"iterations\": 20,\"checkResources\": false,\"maxPingDiff\": 20,\"slaMaxPing\": 100,\"amountDeleteNodes\": 10000,\"speedTest\": false,\"memTest\": true }" > memtest.json

    while [ $iter -lt $iterations ]
    do
     ./swirly "memtest.json"
      #line="$enodes;$fnodes;$memuse"
      #echo $line
      iter=$[$iter+1]
    done

    fnodes=$[$fnodes+50]
  done

  enodes=$[$enodes+50000]
done