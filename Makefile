install: deps
	@go install ./cmd/hdns

test: deps
	@go test -v

deps:
	@go install github.com/coreos/go-etcd/etcd
	@go install github.com/miekg/dns

clean:
	@rm -f $(GOPATH)/bin/hdns
