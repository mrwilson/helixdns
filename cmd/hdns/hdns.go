package main

import (
  "github.com/mrwilson/helixdns"
  "os"
  "os/signal"
  "syscall"
  "log"
)

func main() {
  server := helixdns.Server(9000, "http://localhost:4001/")

  go func() {
    server.Start()
  }()

  sig := make(chan os.Signal)
  signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
  for {
    select {
    case s := <-sig:
      log.Fatalf("Signal (%d) received, stopping\n", s)
    }
  }


}

