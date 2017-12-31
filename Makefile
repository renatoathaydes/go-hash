VERSION=0.5

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
	env GOOS=linux env GOARCH=amd64 go build -ldflags "-s -w" -o release/linux-amd64/$(VERSION)/go-hash	
	env GOOS=linux env GOARCH=386 go build -ldflags "-s -w" -o release/linux-386/$(VERSION)/go-hash	

release-windows:
	env GOOS=windows env GOARCH=amd64 go build -ldflags "-s -w" -o release/windows-amd64/$(VERSION)/go-hash
	env GOOS=windows env GOARCH=386 go build -ldflags "-s -w" -o release/windows-386/$(VERSION)/go-hash

release-darwin:
	env GOOS=darwin env GOARCH=amd64 go build -ldflags "-s -w" -o release/darwin-amd64/$(VERSION)/go-hash

release: test release-linux release-windows release-darwin
