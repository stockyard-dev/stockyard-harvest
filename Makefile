build:
	CGO_ENABLED=0 go build -o harvest ./cmd/harvest/

run: build
	./harvest

test:
	go test ./...

clean:
	rm -f harvest

.PHONY: build run test clean
