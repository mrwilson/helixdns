install: deps
	@go install ./cmd/hdns

test: deps
	@go test -v

deps:
	@go get github.com/coreos/go-etcd/etcd
	@go get github.com/miekg/dns

clean:
	@rm -f $(GOPATH)/bin/hdns
