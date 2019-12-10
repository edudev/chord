#!/usr/bin/env bash

trap 'failure ${LINENO} "$BASH_COMMAND"; exit' INT TERM ERR
trap "echo Exiting...; kill 0" EXIT

failure() {
  local lineno="$1"
  local msg="$2"
  echo "Failed at ${lineno}: ${msg}"
}

THRD=32
CONN=50

NC="$SLURM_NNODES" #Node count: allocated by slurm
IC=$(( NC / 2 )) #Node count: per instance type (memcached or benchmark)

collectd_dir="${HOME}/chord/collectd"
TIMESTAMP="$(date +"%Y%m%d_%H%M%S")"
WORKDIR="${PWD}/results-${TIMESTAMP}-N${NC}-THRD${THRD}-CONN${CONN}"
mkdir -p "$WORKDIR"

#Get hostnames from reserved nodes and put them into an array
mapfile -t NODES < <(srun hostname | sort)

#A is servers', B is clients' array
A=( "${NODES[@]:0:$IC}" )
B=( "${NODES[@]:$IC:$IC}" )

#Remove previously captured collectd data
rm -rf "${HOME}/chord/collectd/var/lib/collectd/"*".cm.cluster"

CHORD_EXEC="${HOME}/opt/bin/chord"
echo $'\n'"Starting Chord..."
CHORD_ARGS=( -N 2 -addr "${A[0]}" )
cat >"chord_0.sh" <<EOT
#!/usr/bin/env bash
"${collectd_dir}/sbin/collectd" -f -C "${collectd_dir}/etc/collectd.conf" &
sleep 5
$CHORD_EXEC ${CHORD_ARGS[@]}
EOT
chmod +x "chord_0.sh"
srun -N1 -w "${A[0]}" "chord_0.sh" >"${WORKDIR}/chord_out0" 2>&1 &
echo "Chord started on: ${A[0]}"$'\n'"Output of Chord: ${WORKDIR}/chord_out0"

(n=1
while [ $n -lt $IC ]; do
CHORD_ARGS=( -N 2 -addr "${A[${n}]}" -join "${A[0]}:21210" )
cat >"chord_${n}.sh" <<EOT
#!/usr/bin/env bash
"${collectd_dir}/sbin/collectd" -f -C "${collectd_dir}/etc/collectd.conf" &
sleep 5
$CHORD_EXEC ${CHORD_ARGS[@]}
EOT
chmod +x "chord_${n}.sh"
srun -N1 -w "${A[${n}]}" "chord_${n}.sh" >"${WORKDIR}/chord_out${n}" 2>&1 &
echo "Chord started on: ${A[${n}]}"$'\n'"Output of Chord: ${WORKDIR}/chord_out${n}"
n=$(( n + 1 ))
done)

STABILISE_EXEC="${HOME}/opt/bin/stabilise"
echo $'\n'"Starting Stabilise..."
STABILISE_ARGS=( "${A[0]}:21210" "$(( IC * 2 ))" )
echo ayyy stabilise args lol "${STABILISE_ARGS[@]}"
srun "-N1" -w "${B[0]}" "$STABILISE_EXEC" "${STABILISE_ARGS[@]}" >"${WORKDIR}/stabilise_out" 2>&1
echo "Stabilise done on: ${B[0]}."

sleep 2

MEMTIER_B_EXEC="${HOME}/opt/bin/memtier_benchmark"
echo $'\n'"Starting Memtier Benchmark..."
(n=0
while [ $n -lt $IC ]; do
MEMTIER_B_ARGS=( -s "${A[${n}]}" -p 11211 -P memcache_text -d 1024 -x 1 -t "$THRD" -c "$CONN" --test-time 10 )
cat >"memtier_${n}.sh" <<EOT
#!/usr/bin/env bash
#hostname
timeout 20s "${collectd_dir}/sbin/collectd" -f -C "${collectd_dir}/etc/collectd.conf" &
sleep 5
$MEMTIER_B_EXEC ${MEMTIER_B_ARGS[@]}
EOT
chmod +x "memtier_${n}.sh"
srun -N1 -w "${B[${n}]}" "memtier_${n}.sh" >"${WORKDIR}/memtier_out${n}" 2>&1 &
echo "Memtier started on: ${B[${n}]}"$'\n'"Benchmark results: ${WORKDIR}/memtier_out${n}"
n=$(( n + 1 ))
done; wait)

sleep 2

n=0
while [ $n -lt $IC ]; do
mv "${HOME}/chord/collectd/var/lib/collectd/${B[${n}]}.cm.cluster/" "${WORKDIR}/collectd_memtier_${n}/"
n=$(( n + 1 ))
done

MEMSLAP_EXEC="${HOME}/opt/bin/memaslap"
echo $'\n'"Starting Memaslap..."
(n=0
while [ $n -lt $IC ]; do
MEMSLAP_ARGS=( -s "${A[${n}]}:11211" -S 1s -T "$THRD" -c "$(( THRD * CONN ))" -t 10s )
cat >"memslap_${n}.sh" <<EOT
#!/usr/bin/env bash
#hostname
timeout 20s "${collectd_dir}/sbin/collectd" -f -C "${collectd_dir}/etc/collectd.conf" &
sleep 5
$MEMSLAP_EXEC ${MEMSLAP_ARGS[@]}
EOT
chmod +x "memslap_${n}.sh"
srun -N1 -w "${B[${n}]}" "memslap_${n}.sh" >"${WORKDIR}/memaslap_out${n}" 2>&1 &
echo "Memaslap started on: ${B[${n}]}"$'\n'"Benchmark results: ${WORKDIR}/memaslap_out${n}"
n=$(( n + 1 ))
done; wait)

sleep 2

n=0
while [ $n -lt $IC ]; do
mv "${HOME}/chord/collectd/var/lib/collectd/${B[${n}]}.cm.cluster/" "${WORKDIR}/collectd_memaslap_${n}/"
n=$(( n + 1 ))
done

echo "Done!"

n=0
while [ $n -lt $IC ]; do
mv "${HOME}/chord/collectd/var/lib/collectd/${A[${n}]}.cm.cluster/" "${WORKDIR}/collectd_memcached_${n}/"
n=$(( n + 1 ))
done

rm memtier_*.sh memslap_*.sh chord_*.sh

exit 0
