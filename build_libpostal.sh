#!/usr/bin/env bash
./bootstrap.sh
mkdir -p /opt/libpostal_data
./configure --datadir=/opt/libpostal_data
echo "#################### Starting make ######################################"
make -j4
echo "#################### Starting make install ##############################"
make install
echo "#################### Starting ldconfig ##################################"
ldconfig || true
