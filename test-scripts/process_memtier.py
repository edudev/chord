#!/usr/bin/env python3


import sys
import csv


def process_memtier(inf, csvoutf):
    for line in inf:
        op, latency, percentile = line.split()

        if op == 'SET':
            continue

        assert op == 'GET'

        latency = float(latency)
        percentile = float(percentile)

        csvoutf.writerow({
            'latency': latency,
            'percentile': percentile,
        })


def main():
    fieldnames = ['latency', 'percentile']
    csvoutf = csv.DictWriter(sys.stdout, fieldnames=fieldnames)
    csvoutf.writeheader()

    process_memtier(sys.stdin, csvoutf)


if __name__ == '__main__':
    main()
