package main

import (
  "github.com/miekg/dns"
  "github.com/coreos/go-etcd/etcd"
  "net"
  "path"
  "strings"
)

type DNSClient interface {
  Lookup(*dns.Msg) (*dns.Msg, error)
  GetAddress() string
}

type ForwardingDNSClient struct {
  Address string
}

func (c ForwardingDNSClient) GetAddress() string {
  return c.Address
}

func (c ForwardingDNSClient) Lookup(req *dns.Msg) (*dns.Msg, error)  {
  return dns.Exchange(req, c.Address)
}

func etcdKeyToDomainName(key string) string {
  name := strings.Split(path.Dir(key),"/")

  domain := make([]string,0)

  for i := len(name)-1; i > 1; i-- {
    domain = append(domain, name[i])
  }

  domain = append(domain, "")

  return strings.Join(domain, ".")
}

func etcdNodeToDnsRecord(node *etcd.Node) []dns.RR {
  switch path.Base(node.Key) {
    case "A":
      header := dns.RR_Header{Name: etcdKeyToDomainName(node.Key), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 5}
      return []dns.RR { &dns.A {Hdr: header, A: net.ParseIP(node.Value)} }
    case "AAAA":
      header := dns.RR_Header{Name: etcdKeyToDomainName(node.Key), Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 5}
      return []dns.RR { &dns.AAAA {Hdr: header, AAAA: net.ParseIP(node.Value)} }
    case "PTR":
      header := dns.RR_Header{Name: etcdKeyToDomainName(node.Key), Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 5}
      return []dns.RR { &dns.PTR {Hdr: header, Ptr: node.Value}}
    case "CNAME":
      header := dns.RR_Header{Name: etcdKeyToDomainName(node.Key), Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 5}
      return []dns.RR { &dns.CNAME {Hdr: header, Target: node.Value}}
    default:
      return nil
  }
}
