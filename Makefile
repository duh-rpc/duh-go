.DEFAULT_GOAL := run
GOLANGCI_LINT = $(GOPATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION = v1.55.2

.PHONY: test
test:
	go test -timeout 20m -v -p 1 -race -parallel=1 ./...

.PHONY: proto
proto:
	# Download and install https://buf.build/ before running
	buf generate

.PHONY: run
run:
	go run -race cmd/demo/main.go

$(GOLANGCI_LINT):
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run
