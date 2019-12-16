#!/bin/bash
set -ve

function run_test {
    nodes="${1}"
    raw_out_file="/tmp/chord_hops_${nodes}.out"
    csv_out_file="chord_hops_${nodes}.csv"

    chord -fix_fingers -learn_nodes -log_hops -addr 172.17.0.1 -N "${nodes}" > "${raw_out_file}" 2>/dev/null &
    stabilise -addr 172.17.0.1:21210 -N "${nodes}" -check_fingers 2> /dev/null

    echo 'Done stabilising'
    start_date=`date --rfc-3339=seconds`
    # docker run --rm memtier_benchmark -n 1000 -x 1 -t 1 -c 50 -P memcache_text -s 172.17.0.1 -p 11211
    memtier_benchmark -n 1000 -x 1 -t 1 -c 50 -P memcache_text -s 172.17.0.1 -p 11211
    kill %1

    ./process_hops.py "${start_date}" "${raw_out_file}" "${csv_out_file}"
}


run_test 8
run_test 16
run_test 32
# run_test 64
# run_test 128
# run_test 256
# run_test 512
# run_test 1024
# run_test 2048
