.PHONY: all
all: test install

ifeq ($(GOPATH),)
  GOPATH=~/go
endif

GODEP=$(GOPATH)/src/github.com/golang/dep/cmd/dep

# get dep if necessary
$(GODEP):
	go get -u github.com/golang/dep/cmd/dep

# download all dependencies into the vendor dir
vendor: $(GODEP) Gopkg.toml Gopkg.lock
	dep ensure

# runs all tests and benchmarks
.PHONY: bench
bench: vendor
	go test -bench .
	cd encryption && go test -bench=.
	cd gohash_db && go test -bench=.

# runs all tests
.PHONY: test
test: vendor
	go test ./...

# installs go-hash
.PHONY: install
install: vendor
	go install

# build a smaller executable without symbols and debug info for all supported OSs and ARCHs
.PHONY: release release-linux release-windows release-darwin

release-linux:
	env GOOS=linux env GOARCH=amd64 go build -ldflags "-s -w" -o releases/go-hash-linux-amd64
	env GOOS=linux env GOARCH=386 go build -ldflags "-s -w" -o releases/go-hash-linux-386

release-windows:
	env GOOS=windows env GOARCH=amd64 go build -ldflags "-s -w" -o releases/go-hash-windows-amd64
	env GOOS=windows env GOARCH=386 go build -ldflags "-s -w" -o releases/go-hash-windows-386

release-darwin:
	env GOOS=darwin env GOARCH=amd64 go build -ldflags "-s -w" -o releases/go-hash-darwin-amd64

release: test release-linux release-windows release-darwin

# clean build artifacts, i.e. everything that is not source code, including the vendor directory.
# Does not remove the installed binary.
.PHONY: clean
clean:
	rm -f go-hash
	rm -rf vendor
	rm -rf releases

# uninstall the go-hash binary
.PHONY: uninstall
uninstall:
	rm -f $(GOPATH)/bin/go-hash
