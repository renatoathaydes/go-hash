all: bench

# runs all tests and benchmarks
bench:
	go test -bench .
	cd encryption && go test -bench=.

# runs all tests
test:
	go test ./...

# build a smaller executable without symbols and debug info
release: test
	go install -ldflags "-s -w"

# build default executable
build:
	go install
