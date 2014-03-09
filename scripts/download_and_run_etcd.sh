#!/bin/sh -x

ETCD_VER=etcd-v0.3.0-linux-amd64

ETCD_URL=https://github.com/coreos/etcd/releases/download/v0.3.0/$ETCD_VER.tar.gz
ETCD_ZIP=/tmp/etcd.tar.gz

echo Cleaning up...
rm -rf $ETCD_ZIP /tmp/$ETCD_VER

echo Downloading etcd...
curl -L $ETCD_URL -o $ETCD_ZIP

echo Unzipping etcd.tar.gz...
tar zxf $ETCD_ZIP -C /tmp

echo Starting etcd...
/tmp/$ETCD_VER/etcd &
