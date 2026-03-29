.PHONY: build clean test

build:
	go build -o sonarqube-cli ./cmd

test:
	go test -v ./...

clean:
	rm -f sonarqube-cli
