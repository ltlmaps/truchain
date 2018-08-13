PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_tools get_vendor_deps build test

get_tools:
	go get github.com/golang/dep/cmd/dep

build:
	go build -o bin/trucli cmd/trucli/main.go && go build -o bin/trucoind cmd/trucoind/main.go

get_vendor_deps:
	@rm -rf vendor/
	@dep ensure

test:
	@go test $(PACKAGES)

benchmark:
	@go test -bench=. $(PACKAGES)

.PHONY: all build test benchmark