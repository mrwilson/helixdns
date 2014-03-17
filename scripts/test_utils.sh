#!/bin/sh

ETCD_URL=http://localhost:4001/v2/keys
ETCD_VER=etcd-v0.2.0-Linux-x86_64

SetEtcdRecord(){
  RECORD=$1
  VALUE=$2
  curl --silent -o /dev/null -XPUT ${ETCD_URL}/${RECORD} -d value="${VALUE}"
}

AssertRecordEquals(){
  ADDRESS=$1
  TYPE=$2
  RECORD=$3

  OUTPUT=$(dig ${ADDRESS} @localhost -p 9000 ${TYPE} +short)

  if [ "${RECORD}" = "${OUTPUT}" ]; then
    echo Successfully resolved ${ADDRESS} to ${RECORD}
  else
    echo Did not successfully resolve ${ADDRESS} to ${RECORD}: got ${OUTPUT} instead
    EXIT_CODE=1
  fi
}
