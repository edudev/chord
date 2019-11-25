#!/bin/bash

#Run first release of chord and test it with memtier and memslap

trap "exit" INT TERM ERR
trap "kill 0" EXIT

TIMESTAMP="$(date +"%Y%m%d_%H%M%S")"

module load prun

CHORD_EXEC="env -i $HOME/remote/chord"
CHORD_ARGS=""
echo "Starting Chord..."
prun -v -np 1 $CHORD_EXEC $CHORD_ARGS < /dev/null >$HOME/chord_out_$TIMESTAMP 2>$HOME/tmpout &
sleep 2

NODE="$(grep -o -P "node.{0,3}" $HOME/tmpout)"
echo "Chord started on $NODE."
echo "Output of Chord: $HOME/chord_out_$TIMESTAMP"

MEMTIER_B_EXEC="$HOME/opt/bin/memtier_benchmark"
#Memtier using 1 thread 1 client
MEMTIER_B_ARGS="-s $NODE -p 11211 -P memcache_text -x 1 -t 1 -c 1"
echo "Running Memtier Benchmark..."
prun -np 1 $MEMTIER_B_EXEC $MEMTIER_B_ARGS </dev/null >$HOME/memtier_out_$TIMESTAMP 2>&1
echo "Benchmark results: $HOME/memtier_out_$TIMESTAMP"

MEMSLAP_EXEC="env -i $HOME/opt/bin/memslap"
MEMSLAP_ARGS="--servers=$NODE"
echo "Running Memslap..."
prun -np 1 $MEMSLAP_EXEC $MEMSLAP_ARGS </dev/null >$HOME/memslap_out_$TIMESTAMP 2>&1
echo "Benchmark results: $HOME/memslap_out_$TIMESTAMP"
echo "Done!"
rm $HOME/tmpout
