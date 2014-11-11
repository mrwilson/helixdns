package main

import (
  "github.com/coreos/go-etcd/etcd"
  "github.com/miekg/dns"
  "log"
  "net"
  "path"
)

type Response interface {
  Value() string
  Node() *etcd.Node
}

type Client interface {
  Get(path string) (Response, error)
  WatchForChanges()
  GetAll(path string) []*etcd.Node
}

type EtcdClient struct {
  InstanceUrl string
  Client      *etcd.Client
}

type EtcdResponse struct {
  Response *etcd.Response
}

func NewEtcdClient(instanceUrl string) Client {
  return &EtcdClient{
    InstanceUrl: instanceUrl,
    Client: etcd.NewClient([]string{instanceUrl}),
  }
}

func (r EtcdResponse) Node() *etcd.Node {
  return r.Response.Node;
}

func (r EtcdResponse) Value() string {
  return r.Response.Node.Value;
}

func validate(node *etcd.Node) (bool, string) {
  recordType := path.Base(node.Key)
  switch recordType {
    case "A":
      return net.ParseIP(node.Value) != nil, "Invalid ip"
    case "CNAME", "PTR":
      return dns.IsFqdn(node.Value), "Domain name not fully-qualified"
    case "SOA":
      return true, ""
    default:
      return false, "Record type not supported"
  }
}

func (c EtcdClient) GetAll(key string) []*etcd.Node {
  response, _ := c.Client.Get(key, false, true)

  nodes := []*etcd.Node{}

  for _, node := range response.Node.Nodes {
    nodes = append(nodes, node)
  }

  return nodes
}

func (c EtcdClient) WatchForChanges() {
  log.Printf("Setting up watch to validate entries")
  channel := make(chan *etcd.Response)

  go func() {
    c.Client.Watch("/helix", 0, true, channel, nil)
  }()

  defer c.catchEtcdPanic()

  for {
    select {
      case resp := <-channel:
        if valid, msg := validate(resp.Node); !valid {
          log.Printf("ERROR: %s (%s => %s)", msg, resp.Node.Key, resp.Node.Value)
        }
    }
  }
}

func (c EtcdClient) catchEtcdPanic() {
  if r := recover(); r != nil {
    log.Fatalf("Panic in setting up watch on /helix. Is etcd running at %s ?", c.InstanceUrl)
  }
}

func (c EtcdClient) Get(path string) (Response, error) {
  resp, err := c.Client.Get(path, false, false)

  if err != nil {
    return nil, err
  }

  return &EtcdResponse{ Response: resp }, nil
}
