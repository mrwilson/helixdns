BATS_DIR := /tmp/bats
ETCD_DIR := /tmp/etcd
ETCD_ADDR := 127.0.0.1:4111

test: go-test integration-test

build: deps
	@go build

go-test: deps
	@go test -v

integration-test: build etcd bats
	@$(ETCD_DIR)/etcd -addr=$(ETCD_ADDR) &
	@sleep 5
	@./helixdns -forward=8.8.8.8:53 -etcd-address=http://$(ETCD_ADDR)/ &
	@sleep 5
	@$(BATS_DIR)/bin/bats tests/ && (killall -9 etcd helixdns || true)

deps:
	@go get github.com/coreos/go-etcd/etcd
	@go get github.com/miekg/dns

etcd:
	@[ -d $(ETCD_DIR) ] || sh ./scripts/download_etcd.sh

bats:
	@[ -d $(BATS_DIR) ] || git clone https://github.com/sstephenson/bats.git $(BATS_DIR)

clean:
	@rm -rf `uname -n`.etcd
	@rm -f helixdns
	@rm -rf $(BATS_DIR) $(ETCD_DIR)
