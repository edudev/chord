#!/usr/bin/env bash

trap 'failure ${LINENO} "$BASH_COMMAND"; exit' INT TERM ERR
trap "echo Exiting...; kill 0" EXIT

failure() {
  local lineno="$1"
  local msg="$2"
  echo "Failed at ${lineno}: ${msg}"
}

NC="$SLURM_NNODES" #Node count: allocated by slurm
IC=$(( NC / 2 )) #Node count: per instance type (memcached or benchmark)

TIMESTAMP="$(date +"%Y%m%d_%H%M%S")"

mapfile -t NODES < <(srun hostname | sort)

MEMCACHED_EXEC="${HOME}/opt/bin/memcached"
MEMCACHED_ARGS=( -v )
echo $'\n'"Starting Memcached..."
OIFS="$IFS"; IFS=","
srun "-N${IC}" -l -w "${NODES[*]:0:$IC}" "$MEMCACHED_EXEC" "${MEMCACHED_ARGS[@]}" >"memcached_out_${TIMESTAMP}" 2>&1 &
echo "Memcached started on nodes: ${NODES[*]:0:$IC}"$'\n'"Output of Memcached: ${PWD}/memcached_out_${TIMESTAMP}"
IFS="$OIFS"
sleep 2

A=( "${NODES[@]:0:$IC}" )
B=( "${NODES[@]:$IC:$IC}" )

MEMTIER_B_EXEC="${HOME}/opt/bin/memtier_benchmark"
echo $'\n'"Starting Memtier Benchmark..."
OIFS="$IFS"; IFS=","
(n=0
while [ $n -lt $IC ]; do
MEMTIER_B_ARGS=( -s "${A[${n}]}" -p 11211 -P memcache_text -x 1 -t 16 -c 50 --test-time 10 )
srun -N1 -w "${B[${n}]}" "$MEMTIER_B_EXEC" "${MEMTIER_B_ARGS[@]}" >"memtier_out${n}_${TIMESTAMP}" 2>&1 &
echo "Memtier started on: ${B[${n}]}"$'\n'"Benchmark results: ${PWD}/memtier_out${n}_${TIMESTAMP}"
n=$(( n + 1 ))
done; wait)
IFS="$OIFS"

MEMSLAP_EXEC="${HOME}/opt/bin/memaslap"
echo $'\n'"Starting Memaslap..."
OIFS="$IFS"; IFS=","
(n=0
while [ $n -lt $IC ]; do
MEMSLAP_ARGS=( -s "${A[${n}]}:11211" -S 1s -T 16 -c 800 -t 10s )
srun -N1 -w "${B[${n}]}" "$MEMSLAP_EXEC" "${MEMSLAP_ARGS[@]}" >"memaslap_out${n}_${TIMESTAMP}" 2>&1 &
echo "Memaslap started on: ${B[${n}]}"$'\n'"Benchmark results: ${PWD}/memaslap_out${n}_${TIMESTAMP}"
n=$(( n + 1 ))
done; wait)
IFS="$OIFS"

echo "Done!"
exit
