#!/bin/sh -x

ETCD_DIR=/tmp/etcd
ETCD_ZIP=/tmp/etcd.tar.gz

ETCD_URL=https://github.com/coreos/etcd/releases/download/v0.3.0/etcd-v0.3.0-linux-amd64.tar.gz

echo Cleaning up...
rm -rf $ETCD_ZIP $ETCD_DIR

echo Downloading etcd...
curl -L $ETCD_URL -o $ETCD_ZIP

echo Unzipping etcd.tar.gz...
mkdir -p $ETCD_DIR
tar zxvf $ETCD_ZIP -C $ETCD_DIR --strip 1
