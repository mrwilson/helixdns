package main

import (
  "encoding/json"
  "github.com/miekg/dns"
  "log"
  "strconv"
  "strings"
  "time"
  "github.com/coreos/go-etcd/etcd"
)

type HelixServer struct {
  Port      int
  Client    Client
  DNSClient DNSClient
}

func ForwardingServer(port int, etcdurl, dnsServerUrl string) HelixServer {
  client := NewEtcdClient(etcdurl)
  dnsClient := ForwardingDNSClient{ Address: dnsServerUrl }

  return HelixServer{ Port: port, Client: client, DNSClient: dnsClient }
}

func Server(port int, etcdurl string) HelixServer {
  client := NewEtcdClient(etcdurl)

  return HelixServer{ Port: port, Client: client }
}

func (s HelixServer) Start() {
  server := &dns.Server{
    Addr:         ":"+strconv.Itoa(s.Port),
    Net:          "udp",
    Handler:      dns.HandlerFunc(s.Handler),
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
  }

  zoneTransferServer := &dns.Server{
    Addr:         ":"+strconv.Itoa(s.Port),
    Net:          "tcp",
    Handler:      dns.HandlerFunc(s.zoneTransferHandler),
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
  }

  go s.Client.WatchForChanges()

  log.Print("Starting servers...")

  go server.ListenAndServe()
  go zoneTransferServer.ListenAndServe()
}

func domainToEtcdKey(domain, record string) string {
  addr := dns.SplitDomainName(domain)
  path := []string{"helix"}

  for i := len(addr) - 1; i >= 0; i-- {
    path = append(path, addr[i])
  }

  if record != "" {
    path = append(path, record)
  }

  return strings.Join(path, "/")
}

func (s HelixServer) getResponse(q dns.Question) (Response, error) {
  value := dns.TypeToString[q.Qtype]
  return s.Client.Get(domainToEtcdKey(q.Name, value))
}

func addNode(records chan dns.RR, node *etcd.Node) {
  if node.Dir {
    for i := range node.Nodes {
      addNode(records, node.Nodes[i])
    }
  } else {
    value := etcdNodeToDnsRecord(node)
    for i := range value {
      records <- value[i]
    }
  }
}

func (s HelixServer) recordsForDomain(domain string) chan dns.RR {

  nodes := s.Client.GetAll(domainToEtcdKey(domain, ""))

  leafNodes := make(chan dns.RR, 0)

  go func() {
    for i := range nodes {
      addNode(leafNodes, nodes[i])
    }
    close(leafNodes)
  }()

  return leafNodes
}

func (s HelixServer) zoneTransferHandler(w dns.ResponseWriter, req *dns.Msg) {

  value := req.Question[0].Qtype

  switch value {
  case dns.TypeAXFR, dns.TypeIXFR:
    domain := req.Question[0].Name

    resp, err := s.getResponse(dns.Question{Qtype:dns.TypeSOA,Name:domain })

    if err != nil {
      log.Printf("Could not find SOA for %s", domain)
      return
    }

    c := make(chan *dns.Envelope)
    transfer := new(dns.Transfer)

    defer close(c)

    err = transfer.Out(w, req, c)

    if err != nil {
      log.Printf("Could not begin zone transfer.")
      return
    }

    var soa *dns.SOA

    err = json.Unmarshal([]byte(resp.Value()), &soa)

    if err != nil {
      log.Printf("Failed to parse SOA record: %s", resp.Value())
      return
    }

    header := dns.RR_Header{Name: domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 5}
    soa.Hdr = header

    records := []dns.RR{soa}

    channel := s.recordsForDomain(domain)
    for record := range channel {
      records = append(records, record)
    }

    records = append(records, soa)

    c <- &dns.Envelope{RR: records}

    w.Hijack()
    return
    default:
      log.Printf("Was not a zone transfer request.")
    }

}

func (s HelixServer) Handler(w dns.ResponseWriter, req *dns.Msg) {
  m := new(dns.Msg)
  m.SetReply(req)

  qType  := req.Question[0].Qtype
  qClass := req.Question[0].Qclass

  header := dns.RR_Header{Name: m.Question[0].Name, Rrtype: qType, Class: qClass, Ttl: 5}
  resp, err := s.getResponse(req.Question[0])

  if err != nil {
    if s.DNSClient != nil {
      log.Printf("Could not get record for %s, forwarding to %s", req.Question[0].Name, s.DNSClient.GetAddress())
      in, _ := s.DNSClient.Lookup(req)
      w.WriteMsg(in)
    } else {
      log.Printf("Could not get record for %s", req.Question[0].Name)
      w.WriteMsg(m)
    }
    return
  }

  switch qType {
    case dns.TypeA, dns.TypeAAAA, dns.TypeCNAME, dns.TypePTR:
      m.Answer = etcdNodeToDnsRecord(resp.Node())
    case dns.TypeSRV:
      var records []SrvRecord
      err := json.Unmarshal([]byte(resp.Value()), &records)
      if err != nil {
        log.Printf("Could not unmarshal SRV record from etcd: %s", resp.Value())
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
    default:
      log.Printf("Unrecognised record type: %d",qType)
  }

  w.WriteMsg(m)
}
