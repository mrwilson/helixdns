## HelixDNS

 [![Build Status](https://travis-ci.org/mrwilson/helixdns.png?branch=master)](https://travis-ci.org/mrwilson/helixdns)

  A simple dns server to read records from etcd. See this [blog post](http://probablyfine.co.uk/2014/03/02/serving-dns-records-from-etcd/) for more information.

## Build Instructions

    go get github.com/mrwilson/helixdns
    make install

## Uses

    hdns -port=<port-to-run-on> -etcd-address=<address-of-etcd-instance>

## SRV Records

  SRV records have multiple pieces of information to serve, so the values stored in etcd under `/helix/com/example/_<protocol>/_<service>/SRV` should be in the form of a list of JSON objects, as below.

    [
      {"Priority":10,"Weight":60,"Port":5060,"Target":"bigbox.example.com."},
      {"Priority":10,"Weight":20,"Port":5060,"Target":"smallbox1.example.com."},
      {"Priority":10,"Weight":10,"Port":5060,"Target":"smallbox2.example.com."},
      {"Priority":10,"Weight":10,"Port":5066,"Target":"smallbox2.example.com."},
      {"Priority":20,"Weight":0, "Port":5060,"Target":"backupbox.example.com."}
    ]

  Setting a SRV record could be done with curl:

    curl -XPUT http://localhost:4001/v2/keys/helix/local/_tcp/_syslog/SRV \
         -d value='[{"Priority":10,"Weight":60,"Port":514,"Target":"syslog.local."}, {"Priority":20,"Weight":10,"Port":514,"Target":"graylog.local."}]'

  The outcome looks like:

    # dig @localhost -p 9000 _syslog._tcp.local SRV +short
    10 60 514 syslog.local.
    20 10 514 graylog.local.
    # dig @localhost -p 9000 _syslog._tcp.local SRV
    ; <<>> DiG 9.9.4-P2-RedHat-9.9.4-11.P2.fc20 <<>> @localhost -p 9000 _syslog._tcp.local SRV
    ; (2 servers found)
    ;; global options: +cmd
    ;; Got answer:
    ;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 44651
    ;; flags: qr rd; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 0
    ;; WARNING: recursion requested but not available
    
    ;; QUESTION SECTION:
    ;_syslog._tcp.local.		IN	SRV
    
    ;; ANSWER SECTION:
    _syslog._tcp.local.	5	IN	SRV	10 60 514 syslog.local.
    _syslog._tcp.local.	5	IN	SRV	20 10 514 graylog.local.
    
    ;; Query time: 1 msec
    ;; SERVER: ::1#9000(::1)
    ;; WHEN: Wed Apr 23 09:19:45 CEST 2014
    ;; MSG SIZE  rcvd: 137

  With python using [pydns](http://sourceforge.net/projects/pydns/):
    
    # cat << EOF > fetch_dns.py
    import DNS
    DNS.ParseResolvConf()
    srv_req = DNS.Request(qtype = 'srv', port=9000, server='localhost')
    srv_result = srv_req.req('_syslog._tcp.local')
    
    for result in srv_result.answers:
        if result['typename'] == 'SRV':
            print result['data']
    EOF
    # python fetch_dns.py
    (10, 60, 514, 'syslog.local')
    (20, 10, 514, 'graylog.local')

## TODO

 * Other types of record that aren't A, AAAA, or SRV.
