#!/bin/sh

EXIT_CODE=0

. scripts/test_utils.sh

/tmp/${ETCD_VER}/etcd &
ETCD_PID=$!

go run cmd/hdns/hdns.go &
SERVER_PID=$!

sleep 5

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
