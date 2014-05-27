package helixdns

import (
  "github.com/coreos/go-etcd/etcd"
  "github.com/miekg/dns"
  "log"
  "net"
  "path"
  "math/rand"
)

type Response interface {
  Value() string
}

type Client interface {
  Get(path string) ([]Response, error)
  WatchForChanges()
}

type EtcdClient struct {
  Client *etcd.Client
}

type EtcdResponse struct {
  Response *etcd.Response
}

func NewEtcdClient(instanceUrl string) Client {
  return &EtcdClient{ Client: etcd.NewClient([]string{instanceUrl}) }
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
    default:
      parentType := path.Base(path.Dir(node.Key))
      if (parentType == "A") {
        return net.ParseIP(node.Value) != nil, "Invalid ip"
      }
	  return false, "Record type not supported"
  }
}

func (c EtcdClient) WatchForChanges() {
  log.Printf("Setting up watch to validate entries")
  channel := make(chan *etcd.Response)

  go func() {
    c.Client.Watch("/helix", 0, true, channel, nil)
  }()

  for {
    select {
      case resp := <-channel:
        if valid, msg := validate(resp.Node); !valid {
          log.Printf("ERROR: %s (%s => %s)", msg, resp.Node.Key, resp.Node.Value)
        }
    }
  }
}

func (c EtcdClient) Get(path string) ([]Response, error) {
  resp, err := c.Client.Get(path, false, true)

  if err != nil {
    return nil, err
  }

  // Check to see if we have a directory instead
  if resp.Node.Nodes != nil {
    ret := make([]Response, len(resp.Node.Nodes))
    for i, node := range resp.Node.Nodes {
      // make a mildly complicated anonymous structure to hold our unwrapped object
      ret[i] = &EtcdResponse{ Response: &etcd.Response{ Node: &etcd.Node{ Value: node.Value}}}
    }
    // shuffle the array
    ret2 := make([]Response, len(resp.Node.Nodes))
    perm := rand.Perm(len(ret2))
    for i, j := range perm {
      ret2[j] = ret[i]
    }
    return ret2, nil
  }
  return []Response{&EtcdResponse{ Response: resp }}, nil
}
