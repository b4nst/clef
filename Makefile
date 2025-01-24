.PHONY: build clean run test

build:
	goreleaser release --snapshot --clean

clean:
	rm -rf dist

run:
	go run cmd/clef/main.go

test:
	go test ./...
