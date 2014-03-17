#!/bin/sh

ETCD_URL=http://localhost:4001/v2/keys
ETCD_VER=etcd-v0.2.0-Linux-x86_64

EXIT_CODE=0

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

/tmp/${ETCD_VER}/etcd &
ETCD_PID=$!

go run cmd/hdns/hdns.go &
SERVER_PID=$!

sleep 10

echo

echo TESTING A RECORD...
SetEtcdRecord      "helix/com/example/A" "123.123.123.123"
AssertRecordEquals "example.com."    "A" "123.123.123.123"

echo TESTING PTR RECORD...
SetEtcdRecord      "helix/arpa/in-addr/123/123/123/123/PTR"  "example.com."
AssertRecordEquals "123.123.123.123.in-addr.arpa."    "PTR"  "example.com."

kill -9 ${SERVER_PID}
kill -9 ${ETCD_PID}

exit ${EXIT_CODE}
