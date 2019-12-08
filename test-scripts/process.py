#!/usr/bin/env python3

import sys
import os
import csv
import statistics


OFFSET_SECS = 1


def to_float(data):
    result = []
    for entry in data:
        for field, val in entry.items():
            entry[field] = float(val)

        result.append(entry)

    return result

def parse_csv(fn):
    with open(fn, newline='') as csvfile:
        reader = csv.DictReader(csvfile)

        data = to_float(reader)
        data = sorted(data, key=lambda x: x['epoch'])

        first_entry = data[0]
        initial_timestamp = first_entry['epoch']
        required_timestamp = initial_timestamp + OFFSET_SECS

        for item in data:
            if item['epoch'] >= required_timestamp:
                return item

    return None


def process_stat_cpu(agg, stat_name, stat_dir):
    files = os.listdir(stat_dir)
    assert len(files) == 1

    item = parse_csv(os.path.join(stat_dir, files[0]))

    if item is None:
        raise ValueError('unable to find cpu value')

    agg[stat_name] = item['value']


def process_stat_load(agg, stat_name, stat_dir):
    pass


def process_stat_memory(agg, stat_name, stat_dir):
    files = os.listdir(stat_dir)
    files = [f for f in files if f.startswith('memory-used-')]
    assert len(files) == 1

    assert 'memory' not in agg

    item = parse_csv(os.path.join(stat_dir, files[0]))

    if item is None:
        raise ValueError('unable to find memory value')

    agg['memory'] = item['value']


def process_stat_interface(agg, stat_name, stat_dir):
    if stat_name != 'interface-eth0':
        return

    files = os.listdir(stat_dir)
    files = [f for f in files if f.startswith('if_octets-')]
    assert len(files) == 1

    assert 'rx' not in agg
    assert 'tx' not in agg

    item = parse_csv(os.path.join(stat_dir, files[0]))

    if item is None:
        raise ValueError('unable to find interface value')

    agg['rx'] = item['rx']
    agg['tx'] = item['tx']


def process_stat(agg, stat_name, stat_dir):
    if stat_name.startswith('cpu-'):
        process_stat_cpu(agg, stat_name, stat_dir)
    elif stat_name.startswith('interface-'):
        process_stat_interface(agg, stat_name, stat_dir)
    elif stat_name == 'load':
        process_stat_load(agg, stat_name, stat_dir)
    elif stat_name == 'memory':
        process_stat_memory(agg, stat_name, stat_dir)
    else:
        raise ValueError('Unknown stat', stat_name)


def process_node(node_name, node_dir):
    agg = {}
    for stat_name in os.listdir(node_dir):
        process_stat(agg, stat_name, os.path.join(node_dir, stat_name))

    cpu_data = [val for key, val in agg.items() if key.startswith('cpu-')]

    agg['cpu-max'] = max(cpu_data)
    agg['cpu-avg'] = statistics.mean(cpu_data)

    return agg


def aggregate_stats(data, func):
    result = {}
    for field, vals in data.items():
        result[field] = func(vals)

    return result


def process_stats(data_dir):
    all_stats = {
        'memory': [],
        'rx': [],
        'tx': [],
        'cpu-max': [],
        'cpu-avg': [],
        # 'cpu': [],
    }

    with open('stats.csv', 'w', newline='') as csvfile:
        fieldnames = ['node', 'memory', 'rx', 'tx', 'cpu-max', 'cpu-avg']
        fieldnames.extend('cpu-' + str(i) for i in range(0, 32))
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        writer.writeheader()

        for node_name in os.listdir(data_dir):
            stats = process_node(node_name, os.path.join(data_dir, node_name),)
            stats['node'] = node_name

            writer.writerow(stats)

            all_stats['memory'].append(stats['memory'])
            all_stats['rx'].append(stats['rx'])
            all_stats['tx'].append(stats['tx'])
            all_stats['cpu-max'].append(stats['cpu-max'])
            all_stats['cpu-avg'].append(stats['cpu-avg'])

            # for i in range(0, 32):
            #     all_stats['cpu'].append(stats['cpu-' + str(i)])

        stats_max = aggregate_stats(all_stats, max)
        stats_max['node'] = 'aggregated.max'
        writer.writerow(stats_max)

        stats_avg = aggregate_stats(all_stats, statistics.mean)
        stats_avg['node'] = 'aggregated.avg'
        writer.writerow(stats_avg)


def main():
    data_dir = sys.argv[1]
    process_stats(data_dir)


if __name__ == '__main__':
    main()
