#!/usr/bin/env python3

import re
import sys
import csv
import pandas as pd



def read_df(old_df, fn):
    nodes_match = re.search('_(\d+).csv', fn)
    assert nodes_match

    nodes = int(nodes_match.group(1))

    new_df = pd.read_csv(fn)
    if old_df is None:
        return new_df.rename(columns={'hops': str(nodes)})
        return new_df

    old_df[str(nodes)] = new_df['hops']
    return old_df


def parse_boxplot(box):
    for whisker_lower, whisker_upper, boxrect, boxmedian, boxmean in \
        zip(box['whiskers'][::2], box['whiskers'][1::2], box['boxes'], box['medians'], box['means']):

        whisker_lower = whisker_lower.get_ydata()[1]
        whisker_upper = whisker_upper.get_ydata()[1]

        assert boxrect.get_ydata()[0] == boxrect.get_ydata()[1]
        assert boxrect.get_ydata()[2] == boxrect.get_ydata()[3]
        assert boxrect.get_ydata()[0] == boxrect.get_ydata()[4]

        boxrect_lower = boxrect.get_ydata()[0]
        boxrect_upper = boxrect.get_ydata()[3]

        assert boxmedian.get_ydata()[0] == boxmedian.get_ydata()[1]

        median = boxmedian.get_ydata()[0]
        mean = boxmean.get_ydata()[0]

        yield {
            'whisker_lower': whisker_lower,
            'whisker_upper': whisker_upper,
            'boxrect_lower': boxrect_lower,
            'boxrect_upper': boxrect_upper,
            'median': median,
            'mean': mean,
        }


def combine_hops(hops_in_fns, hops_out_fn):
    hop_counts = None
    for hops_in_fn in hops_in_fns:
        hop_counts = read_df(hop_counts, hops_in_fn)

    box = hop_counts.plot(kind='box', whis=1.5, showmeans=True, notch=False,
                          showfliers=True, return_type='dict')

    fieldnames = [
        'N',
        'whisker_lower',
        'whisker_upper',
        'boxrect_lower',
        'boxrect_upper',
        'median',
        'mean',
    ]

    with open(hops_out_fn, 'w', newline='') as outf:
        csvoutf = csv.DictWriter(outf, fieldnames=fieldnames)
        csvoutf.writeheader()

        for N, entry in zip(hop_counts, parse_boxplot(box)):
            entry['N'] = N
            csvoutf.writerow(entry)


def main():
    hops_out_fn = sys.argv[1]
    hops_in_fns = sys.argv[2:]

    combine_hops(hops_in_fns, hops_out_fn)


if __name__ == '__main__':
    main()
