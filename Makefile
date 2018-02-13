build:
	go build
	go vet
	golint

install:
	go install

clean:
	go clean

fmt:
	gofmt -w *.go
