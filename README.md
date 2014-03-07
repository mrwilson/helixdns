## HelixDNS

 [![Build Status](https://travis-ci.org/mrwilson/helixdns.png?branch=master)](https://travis-ci.org/mrwilson/helixdns)

  A simple dns server to read records from etcd. See this [blog post](http://probablyfine.co.uk/2014/03/02/serving-dns-records-from-etcd/) for more information.

## Build Instructions

    go get github.com/mrwilson/helixdns
    make install

## Uses

    hdns -port=<port-to-run-on> -etcd-address=<address-of-etcd-instance>

## TODO

 * Other types of record that aren't A or AAAA.
