.PHONY: all
all: test

# runs all tests and benchmarks
.PHONY: bench
bench:
	go test -bench .
	cd encryption && go test -bench=.

# runs all tests
.PHONY: test
test:
	go test ./...

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
