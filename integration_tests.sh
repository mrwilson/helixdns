#!/bin/sh

ETCD_URL=http://localhost:4001/v2/keys

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
    exit 1
  fi
}

go run cmd/hdns/hdns.go &
SERVER_PID=$!
sleep 5

echo

echo TESTING A RECORD...
SetEtcdRecord      "helix/com/example/A" "123.123.123.123"
AssertRecordEquals "example.com."    "A" "123.123.123.123"

kill -9 ${SERVER_PID}
killall -9 etcd
