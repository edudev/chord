#!/bin/bash
trap "exit" INT TERM ERR
trap "kill 0" EXIT

TIMESTAMP="$(date +"%Y%m%d_%H%M%S")"

module load prun

MEMCACHED_EXEC="$HOME/opt/bin/memcached"
MEMCACHED_ARGS=""

echo "Starting Memcached..."
prun -v -np 1 $MEMCACHED_EXEC $MEMCACHED_ARGS </dev/null >$HOME/memcached_out_$TIMESTAMP 2>$HOME/tmpout &
sleep 2 #try to increase this if nothing happens after previous echo

NODE="$(grep -o -P "node.{0,3}" $HOME/tmpout)"
echo "Memcached started on $NODE."
echo "Output of Memcached: $HOME/memcached_out_$TIMESTAMP"

MEMTIER_B_EXEC="$HOME/opt/bin/memtier_benchmark"
MEMTIER_B_ARGS="-s $NODE -p 11211 -P memcache_text -x 1"
echo "Running Memtier Benchmark..."
prun -np 1 $MEMTIER_B_EXEC $MEMTIER_B_ARGS </dev/null >$HOME/memtier_out_$TIMESTAMP 2>&1
echo "Benchmark results: $HOME/memtier_out_$TIMESTAMP"

MEMSLAP_EXEC="$HOME/opt/bin/memslap"
MEMSLAP_ARGS="--servers=$NODE"
echo "Running Memslap..."
prun -np 1 $MEMSLAP_EXEC $MEMSLAP_ARGS </dev/null >$HOME/memslap_out_$TIMESTAMP 2>&1
echo "Benchmark results: $HOME/memslap_out_$TIMESTAMP"

echo "Done!"

rm $HOME/tmpout
