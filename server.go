package helixdns

import (
  "encoding/json"
  "github.com/miekg/dns"
  "log"
  "net"
  "strconv"
  "strings"
)

type HelixServer struct {
  Port      int
  Client    Client
  DNSClient DNSClient
}

func ForwardingServer(port int, etcdurl, dnsServerUrl string) *HelixServer {
  client := NewEtcdClient(etcdurl)
  dnsClient := ForwardingDNSClient{ Address: dnsServerUrl }

  return &HelixServer{ Port: port, Client: client, DNSClient: dnsClient }
}

func Server(port int, etcdurl string) *HelixServer {
  client := NewEtcdClient(etcdurl)

  return &HelixServer{ Port: port, Client: client }
}

func (s HelixServer) Start() {
  server := &dns.Server{
    Addr:         ":"+strconv.Itoa(s.Port),
    Net:          "udp",
    Handler:      dns.HandlerFunc(s.Handler),
    ReadTimeout:  10,
    WriteTimeout: 10,
  }

  go s.Client.WatchForChanges()

  log.Print("Starting server...")

  server.ListenAndServe()
}

func (s HelixServer) getResponse(q dns.Question) ([]Response, error) {
  addr := dns.SplitDomainName(q.Name)
  path := []string{"helix"}

  for i := len(addr) - 1; i >= 0; i-- {
    path = append(path, addr[i])
  }

  path = append(path, dns.TypeToString[q.Qtype])

  return s.Client.Get(strings.Join(path, "/"))
}

func (s HelixServer) Handler(w dns.ResponseWriter, req *dns.Msg) {
  m := new(dns.Msg)
  m.SetReply(req)

  qType  := req.Question[0].Qtype
  qClass := req.Question[0].Qclass

  resp, err := s.getResponse(req.Question[0])

  if err != nil {
    if s.DNSClient != nil {
      log.Printf("Could not get record for %s, forwarding to %s", req.Question[0].Name, s.DNSClient.GetAddress())
      in, _ := s.DNSClient.Lookup(req)
      w.WriteMsg(in)
      return
    } else {
      // We need to check to see if we were asked for an A, but we don't have an A, but a CNAME instead
      if qType == dns.TypeA {
        // Current theory is to check to see since we don't have an A, reprocess the request with a CNAME instead
        req.Question[0].Qtype = dns.TypeCNAME
        qType = dns.TypeCNAME
        resp, err = s.getResponse(req.Question[0])
        // If we still don't, bail
        if err != nil {
          log.Printf("Could not get CNAME for %s either (%s)", req.Question[0].Name, err)
          w.WriteMsg(m)
          return
        }
      } else {
        log.Printf("Could not get record for %s (%s)", req.Question[0].Name, err)
        w.WriteMsg(m)
        return
      }
    }
  }

  header := dns.RR_Header{Name: m.Question[0].Name, Rrtype: qType, Class: qClass, Ttl: 5}

  switch qType {
    case dns.TypeA:
      m.Answer = make([]dns.RR, len(resp))
      for i, node := range resp {
        m.Answer[i] = &dns.A {Hdr: header, A: net.ParseIP(node.Value())}
      }
    case dns.TypeAAAA:
      m.Answer = make([]dns.RR, 1)
      m.Answer[0] = &dns.AAAA {Hdr: header, AAAA: net.ParseIP(resp[0].Value())}
    case dns.TypeSRV:
      var records []SrvRecord
      err := json.Unmarshal([]byte(resp[0].Value()), &records)
      if err != nil {
        log.Printf("Could not unmarshal SRV record from etcd: %s", resp[0].Value())
      } else {
        m.Answer = make([]dns.RR, len(records))
        for i := range records {
          m.Answer[i] = &dns.SRV {
            Hdr:      header,
            Priority: records[i].Priority,
            Weight:   records[i].Weight,
            Port:     records[i].Port,
            Target:   records[i].Target,
          }
        }
      }
    case dns.TypePTR:
      m.Answer = make([]dns.RR, 1)
      m.Answer[0] = &dns.PTR {Hdr: header, Ptr: resp[0].Value()}
    case dns.TypeCNAME:
      m.Answer = make([]dns.RR, 1)
      m.Answer[0] = &dns.CNAME {Hdr: header, Target: resp[0].Value()}
    default:
      log.Printf("Unrecognised record type: %d",qType)
  }

  w.WriteMsg(m)
}
