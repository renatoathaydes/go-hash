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
.PHONY: check-version release release-linux release-windows release-darwin

check-version:
ifndef VERSION
	$(error VERSION not set. Run make with 'VERSION=x.x.x make ...')
endif
	@:

release-linux: check-version
	env GOOS=linux env GOARCH=amd64 go build -ldflags "-s -w" -o releases/linux-amd64/$(VERSION)/go-hash	
	env GOOS=linux env GOARCH=386 go build -ldflags "-s -w" -o releases/linux-386/$(VERSION)/go-hash	

release-windows: check-version
	env GOOS=windows env GOARCH=amd64 go build -ldflags "-s -w" -o releases/windows-amd64/$(VERSION)/go-hash
	env GOOS=windows env GOARCH=386 go build -ldflags "-s -w" -o releases/windows-386/$(VERSION)/go-hash

release-darwin: check-version
	env GOOS=darwin env GOARCH=amd64 go build -ldflags "-s -w" -o releases/darwin-amd64/$(VERSION)/go-hash

release: test release-linux release-windows release-darwin
