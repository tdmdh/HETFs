.PHONY: build test run clean

build:
	go build -o bin/halalctl .

test:
	go test ./... -v -count=1

run:
	go run . $(ARGS)

clean:
	rm -rf bin/ halalctl.db
