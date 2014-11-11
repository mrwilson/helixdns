#!/bin/sh

ETCD_CTL=/tmp/etcd/etcdctl

set_etcd_record (){
  RECORD=$1
  VALUE=$2
  ${ETCD_CTL} --peers 127.0.0.1:4111 set /helix/${RECORD} "${VALUE}" > /dev/null
}

dig_record (){
  ADDRESS=$1
  TYPE=$2
  dig ${ADDRESS} @localhost -p 9000 ${TYPE} +short
}

dig_record_contains () {
  ADDRESS=$1
  TYPE=$2
  SEARCH=$3
  dig ${ADDRESS} @localhost -p 9000 ${TYPE} +short | grep "${SEARCH}"
}
