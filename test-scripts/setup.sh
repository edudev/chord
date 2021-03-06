#!/usr/bin/env bash

set -ve

#Install script for DAS5

#GO_VER="1.13.4.linux-amd64"
MEMCACHED_VER="1.5.20"
MEMTIER_B_VER="1.2.17"
LMEMCACHED_VER="1.0.18"

mkdir -p "${HOME}/opt/"
#mkdir -p "${HOME}/go/src/github.com/edudev"

#cd "$HOME"
#wget https://dl.google.com/go/go${GO_VER}.tar.gz
#tar -C "${HOME}/opt" -xzf "go${GO_VER}.tar.gz"
#export PATH="${PATH}:${HOME}/opt/go/bin"
#cd "${HOME}/go/src/github.com/edudev"
#git clone https://github.com/edudev/chord.git
#cd "$HOME"
#ln -s "${HOME}/go/src/github.com/edudev/chord" remote/
#rm -rf "go${GO_VER}.tar.gz"

cd "$HOME"
wget "http://www.memcached.org/files/memcached-${MEMCACHED_VER}.tar.gz"
tar -zxf "memcached-${MEMCACHED_VER}.tar.gz"
cd "memcached-${MEMCACHED_VER}"
./configure "--prefix=${HOME}/opt/" &&
        make &&
        make test &&
        make install
cd "$HOME"
rm -rf "memcached-${MEMCACHED_VER}"
rm "memcached-${MEMCACHED_VER}.tar.gz"

cd "$HOME"
wget "https://github.com/RedisLabs/memtier_benchmark/archive/${MEMTIER_B_VER}.tar.gz"
tar -zxf "${MEMTIER_B_VER}.tar.gz"
cd "memtier_benchmark-${MEMTIER_B_VER}"
autoreconf -ivf &&
        ./configure "--prefix=${HOME}/opt/" &&
        make &&
        make install-exec
        make install-data || true
cd "$HOME"
rm -rf "memtier_benchmark-${MEMTIER_B_VER}"
rm "${MEMTIER_B_VER}.tar.gz"

cd "$HOME"
wget "https://launchpad.net/libmemcached/1.0/${LMEMCACHED_VER}/+download/libmemcached-${LMEMCACHED_VER}.tar.gz"
tar -zxf "libmemcached-${LMEMCACHED_VER}.tar.gz"
cd "libmemcached-${LMEMCACHED_VER}"
./configure --enable-memaslap "--prefix=${HOME}/opt/" &&
        make &&
        make install
cd "$HOME"
rm -rf "libmemcached-${LMEMCACHED_VER}"
rm "libmemcached-${LMEMCACHED_VER}.tar.gz"

"$(dirname ${0})/collectd/setup.sh"

exit 0
