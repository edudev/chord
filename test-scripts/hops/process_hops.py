#!/usr/bin/env python3

import sys
from datetime import datetime
import csv


def process_hops(start_datetime, hops_in_fn, hops_out_fn):
    start_datetime = datetime.strptime(start_datetime, '%Y-%m-%d %H:%M:%S%z')

    with open(hops_out_fn, 'w', newline='') as outf:
        csvoutf = csv.writer(outf)
        csvoutf.writerow(['hops'])

        with open(hops_in_fn) as f:
            for line in f:
                timestamp, hops = line.split()
                timestamp = datetime.strptime(timestamp, '%Y-%m-%dT%H:%M:%S%z')

                if timestamp < start_datetime:
                    continue

                hops = int(hops)
                csvoutf.writerow([hops])


def main():
    assert len(sys.argv) == 4
    start_datetime = sys.argv[1]
    hops_in_fn = sys.argv[2]
    hops_out_fn = sys.argv[3]

    process_hops(start_datetime, hops_in_fn, hops_out_fn)

if __name__ == '__main__':
    main()
