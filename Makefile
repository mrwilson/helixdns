install: deps
	@go install ./cmd/hdns

deps:
	@go install github.com/coreos/go-etcd/etcd
	@go install github.com/miekg/dns

clean:
	@rm -f $(GOPATH)/bin/hdns
