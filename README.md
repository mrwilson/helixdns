## HelixDNS

  A simple dns server to read records from etcd.

## Build Instructions

    go get github.com/mrwilson/helixdns

## Uses

    go run ./cmd/hdns/hdns.go

## Requirements

  A running etcd server is required. HelixDNS defaults to http://localhost:4001/, the etcd default.

  Records are stored as keys under /helix in etcd's key-value store. So the A record for foo.example.com. would be stored as the value in the node at

    /helix/com/example/foo/A

## TODO

 * Literally every other type of record that's not A.
 * Tests.
