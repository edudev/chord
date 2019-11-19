#!/bin/sh

MEMCACHED_VER="1.5.20"
MEMTIER_B_VER="1.2.17"

mkdir -p $HOME/opt/

cd $HOME
wget http://www.memcached.org/files/memcached-$MEMCACHED_VER.tar.gz
tar -zxf memcached-$MEMCACHED_VER.tar.gz
cd memcached-$MEMCACHED_VER/
./configure --prefix=$HOME/opt/ &&
	make &&
	make test &&
	make install
cd ..
rm -rf memcached-$MEMCACHED_VER/
rm memcached-$MEMCACHED_VER.tar.gz

cd $HOME
wget https://github.com/RedisLabs/memtier_benchmark/archive/$MEMTIER_B_VER.tar.gz
tar -zxf $MEMTIER_B_VER.tar.gz
cd memtier_benchmark-$MEMTIER_B_VER/
autoreconf -ivf &&
	./configure --prefix=$HOME/opt/ &&
	make &&
	make install-exec
cd ..
rm -rf memtier_benchmark-$MEMTIER_B_VER/
rm $MEMTIER_B_VER.tar.gz

