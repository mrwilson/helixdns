package helixdns

import (
  "github.com/coreos/go-etcd/etcd"
  "github.com/miekg/dns"
  "log"
  "net"
  "path"
)

type Response interface {
  Value() string
}

type Client interface {
  Get(path string) (Response, error)
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

func (c EtcdClient) Get(path string) (Response, error) {
  resp, err := c.Client.Get(path, false, false)

  if err != nil {
    return nil, err
  }

  return &EtcdResponse{ Response: resp }, nil
}
