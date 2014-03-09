install: deps
	@go install ./cmd/hdns

test: deps
	@go test -v

integration-test: deps
	@sh ./scripts/download_and_run_etcd.sh
	@sh ./integration_tests.sh

deps:
	@go get github.com/coreos/go-etcd/etcd
	@go get github.com/miekg/dns

clean:
	@rm -f $(GOPATH)/bin/hdns
