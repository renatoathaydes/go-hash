.PHONY: all
all: test install

ifeq ($(GOPATH),)
  GOPATH=~/go
endif

# setup dependencies
.PHONY: setup
setup:
	go get

# runs all tests and benchmarks
.PHONY: bench
bench: setup
	go test -bench .
	cd encryption && go test -bench=.
	cd gohash_db && go test -bench=.

# runs all tests
.PHONY: test
test: setup
	go test ./...

# installs go-hash
.PHONY: install
install: setup
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
	env GOOS=darwin env GOARCH=arm64 go build -ldflags "-s -w" -o releases/go-hash-darwin-arm64

release: test release-linux release-windows release-darwin

# clean build artifacts, i.e. everything that is not source code.
# Does not remove the installed binary.
.PHONY: clean
clean:
	rm -f go-hash
	rm -rf releases

# uninstall the go-hash binary
.PHONY: uninstall
uninstall:
	rm -f $(GOPATH)/bin/go-hash
