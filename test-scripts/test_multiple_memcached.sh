#!/bin/bash
trap "kill 0" EXIT

TIMESTAMP="$(date +"%Y%m%d_%H%M%S")"

module load prun

MEMCACHED_EXEC="env -i $HOME/opt/bin/memcached"
MEMCACHED_ARGS=""
echo "Starting Memcached..."
prun -v -np 4 $MEMCACHED_EXEC $MEMCACHED_ARGS < /dev/null >$HOME/memcached_out_$TIMESTAMP 2>$HOME/tmpout &
sleep 4

echo "Memcached started on nodes:"
i=0
for mLine in `grep -o -P "node.{0,3}" $HOME/tmpout`
do
  r[i]=${mLine}
  echo ${r[i]}
  ((i++))
done
echo "Output of Memcached: $HOME/memcached_out_$TIMESTAMP"

MEMSLAP_EXEC="env -i $HOME/opt/bin/memslap"
MEMSLAP_ARGS="--servers=${r[0]},${r[1]},${r[2]},${r[3]}"
echo $MEMSLAP_ARGS
echo "Running Memslap..."
prun -np 1 $MEMSLAP_EXEC $MEMSLAP_ARGS </dev/null >$HOME/memslap_out_$TIMESTAMP 2>&1
echo "Benchmark results: $HOME/memslap_out_$TIMESTAMP"

rm $HOME/tmpout
