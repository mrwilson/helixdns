package helixdns

import (
  "github.com/miekg/dns"
  "github.com/coreos/go-etcd/etcd"
  "log"
  "net"
  "strconv"
  "strings"
)

type HelixServer struct {
  Port   int
  Client *etcd.Client
}

func Server(port int, etcdurl string) *HelixServer {
  client := etcd.NewClient([]string{ etcdurl  })
  return &HelixServer {
    Port: port,
    Client: client,
  }
}

func (s HelixServer) Start() {
  handler := newHandler(s.Client)
  server := &dns.Server{
    Addr:         ":"+strconv.Itoa(s.Port),
    Net:          "udp",
    Handler:      dns.HandlerFunc(handler),
    ReadTimeout:  10,
    WriteTimeout: 10,
  }

  log.Print("Starting server...")

  server.ListenAndServe()
}

func getResponse(client *etcd.Client, q dns.Question) (*etcd.Response, error) {
  addr := dns.SplitDomainName(q.Name)
  path := []string{"helix"}

  for s := range addr {
    path = append(path, addr[len(addr)-s-1])
  }

  path = append(path, dns.TypeToString[q.Qtype])

  return client.Get(strings.Join(path, "/"), false, false)
}

func newHandler(client *etcd.Client) func(dns.ResponseWriter, *dns.Msg) {
  return func (w dns.ResponseWriter, req *dns.Msg) {
    m := new(dns.Msg)
    m.SetReply(req)

    qType  := req.Question[0].Qtype
    qClass := req.Question[0].Qclass

    header := dns.RR_Header{Name: m.Question[0].Name, Rrtype: qType, Class: qClass, Ttl: 5}

    resp, err := getResponse(client, req.Question[0])

    if err == nil && qType == dns.TypeA {
      m.Answer = make([]dns.RR, 1)
      m.Answer[0] = &dns.A {Hdr: header, A: net.ParseIP(resp.Node.Value)}
    }

    w.WriteMsg(m)
  }
}
