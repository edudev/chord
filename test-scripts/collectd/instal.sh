#!/bin/bash
set -ve

wget -N https://github.com/collectd/collectd/releases/download/5.10.0/collectd-5.10.0.tar.bz2
tar -xvf collectd-5.10.0.tar.bz2
collectd_dir="${HOME}/chord/collectd/"
cd collectd-5.10.0
./configure --prefix="${collectd_dir}" --enable-csv
make
make install

sed -i.bak \
    -e 's/${exec_prefix}/${prefix}/g' \
    -e "s/\${prefix}/${collectd_dir//\//\\/}/g" "${collectd_dir}/etc/collectd.conf" \
    -e 's/^#\(BaseDir\|PIDFile\|PluginDir\|TypesDB\)/\1/g' \
    -e 's/^\(LoadPlugin syslog\)/#\1/g' \
    -e 's/^#*\(LoadPlugin logfile\)/\1/g' \
    -e 's/^#*\(LoadPlugin csv\)/\1/g'

"${collectd_dir}/sbin/collectd" -f -C "${collectd_dir}/etc/collectd.conf"
