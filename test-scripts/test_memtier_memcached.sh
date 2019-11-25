#!/bin/bash
trap "kill 0" EXIT

TIMESTAMP="$(date +"%Y%m%d_%H%M%S")"

module load prun

MEMCACHED_EXEC="env -i $HOME/opt/bin/memcached"
MEMCACHED_ARGS=""
echo "Starting Memcached..."
prun -v -np 4 $MEMCACHED_EXEC $MEMCACHED_ARGS < /dev/null >$HOME/memcached_out_$TIMESTAMP 2>$HOME/tmpout &
sleep 30

echo "Memcached started on nodes:"
i=0
for mLine in `grep -o -P "node.{0,3}" $HOME/tmpout`
do
  r[i]=${mLine}
  echo ${r[i]}
  ((i++))
done
echo "Output of Memcached: $HOME/memcached_out_$TIMESTAMP"
echo "Launching benchmarks"
MEMTIER_B_EXEC="$HOME/opt/bin/memtier_benchmark"
for n in "${r[@]}"
do
MEMTIER_B_ARGS="-s $n -p 11211 -P memcache_text -x 1"
prun -np 4 $MEMTIER_B_EXEC $MEMTIER_B_ARGS </dev/null >$HOME/memtier_out_$TIMESTAMP$n 2>&1
echo "Running Memtier Benchmark on $n ..."
echo "Benchmark results: $HOME/memtier_out_$TIMESTAMP$n"
done
rm $HOME/tmpout
echo "Done!"
