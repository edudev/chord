#!/bin/bash
set -ve

wget -N https://download.savannah.gnu.org/releases/freetype/freetype-2.10.1.tar.gz
tar -xvf freetype-2.10.1.tar.gz
freetype2_dir="${HOME}/chord/freetype2/"
cd freetype-2.10.1/
./configure --prefix="${freetype2_dir}"
make
make install
export PKG_CONFIG_PATH="${freetype2_dir}/lib/pkgconfig/:${PKG_CONFIG_PATH}"
export CPPFLAGS="-I${freetype2_dir}/include/ ${CPPFLAGS}"
export LDFLAGS="-L${freetype2_dir}/lib/ ${LDFLAGS}"
cd ../

wget -N https://www.cairographics.org/releases/pixman-0.38.4.tar.gz
tar -xvf pixman-0.38.4.tar.gz
pixman_dir="${HOME}/chord/pixman/"
cd pixman-0.38.4/
./configure --prefix="${pixman_dir}"
make
make install
export PKG_CONFIG_PATH="${pixman_dir}/lib/pkgconfig/:${PKG_CONFIG_PATH}"
export CPPFLAGS="-I${pixman_dir}/include/ ${CPPFLAGS}"
export LDFLAGS="-L${pixman_dir}/lib/ ${LDFLAGS}"
cd ../

wget -N https://www.cairographics.org/releases/cairo-1.16.0.tar.xz
tar -xvf cairo-1.16.0.tar.xz
cairo_dir="${HOME}/chord/cairo/"
cd cairo-1.16.0/
./configure --prefix="${cairo_dir}"
make
make install
export PKG_CONFIG_PATH="${cairo_dir}/lib/pkgconfig/:${PKG_CONFIG_PATH}"
export CPPFLAGS="-I${cairo_dir}/include/cairo/ ${CPPFLAGS}"
export LDFLAGS="-L${cairo_dir}/lib/ ${LDFLAGS}"
cd ../

wget -N http://ftp.gnome.org/pub/GNOME/sources/pango/1.28/pango-1.28.4.tar.gz
tar -xvf pango-1.28.4.tar.gz
pango_dir="${HOME}/chord/pango/"
cd pango-1.28.4/
./configure --prefix="${pango_dir}"
make
make install
export PKG_CONFIG_PATH="${pango_dir}/lib/pkgconfig/:${PKG_CONFIG_PATH}"
export CPPFLAGS="-I${pango_dir}/include/pango-1.0/ ${CPPFLAGS}"
export LDFLAGS="-L${pango_dir}/lib/ ${LDFLAGS}"
cd ../


# wget ftp://ftp.pcre.org/pub/pcre/pcre-8.43.zip
# unzip pcre-8.43.zip
# 
# pcre_dir="${HOME}/chord/pcre/"
# 
# cd pcre-8.43
# ./configure --prefix="${pcre_dir}"
# make
# make install

wget -N https://oss.oetiker.ch/rrdtool/pub/rrdtool-1.7.2.tar.gz
tar -xvf rrdtool-1.7.2.tar.gz
rrdtool_dir="${HOME}/chord/rrdtool/"
cd rrdtool-1.7.2
./configure --prefix="${rrdtool_dir}"
make
make install
export PKG_CONFIG_PATH="${rrdtool_dir}/lib/pkgconfig/:${PKG_CONFIG_PATH}"
export CPPFLAGS="-I${rrdtool_dir}/include/ ${CPPFLAGS}"
export LDFLAGS="-L${rrdtool_dir}/lib/ ${LDFLAGS}"
cd ../


wget -N https://github.com/collectd/collectd/releases/download/5.10.0/collectd-5.10.0.tar.bz2
tar -xvf collectd-5.10.0.tar.bz2
collectd_dir="${HOME}/chord/collectd/"
cd collectd-5.10.0
./configure --prefix="${collectd_dir}" --enable-rrdtool
make
make install

sed -i.bak \
    -e 's/${exec_prefix}/${prefix}/g' \
    -e "s/\${prefix}/${collectd_dir//\//\\/}/g" "${collectd_dir}/etc/collectd.conf" \
    -e 's/^#\(BaseDir\|PIDFile\|PluginDir\|TypesDB\)/\1/g' \
    -e 's/^\(LoadPlugin syslog\)/#\1/g' \
    -e 's/^#*\(LoadPlugin logfile\)/\1/g'


"${collectd_dir}/sbin/collectd" -f -C "${collectd_dir}/etc/collectd.conf"
