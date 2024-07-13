GITCOMMIT := $(shell git rev-parse HEAD)
GITDATE := $(shell git show -s --format='%ct')

LDFLAGSSTRING +=-X main.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X main.GitDate=$(GITDATE)
LDFLAGS := -ldflags "$(LDFLAGSSTRING)"

eth-wallet:
	env GO111MODULE=on go build -v $(LDFLAGS) ./cmd/eth-wallet

clean:
	rm eth-wallet

test:
	go test -v ./...

lint:
	golangci-lint run ./...

.PHONY: \
	eth-wallet \
	bindings \
	bindings-scc \
	clean \
	test \
	lint