# Makefile

export CGO_ENABLED=0

# Set the output directory for the binaries
OUTPUT_DIR := bin

GO_SOURCES = $(wildcard internal/*/*.go internal/*.go)
OUTPUT_BINS = $(patsubst cmd/%, bin/%, $(wildcard cmd/*))

build: $(OUTPUT_BINS)

bin/%: ./cmd/%/main.go $(GO_SOURCES)
	GOOS=linux GOARCH=amd64 go build -o $@.linux.amd64 ./$(shell dirname $<)
	GOOS=windows GOARCH=amd64 go build -o $@.win.amd64.exe ./$(shell dirname $<)

test:
	go test -v -cover ./...

clean:
	-rm -rf bin/
