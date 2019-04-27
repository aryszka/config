default: build

deps:
	go get github.com/aryszka/treerack/...

build:
	go build

check:
	treerack check-syntax ini/syntax.treerack
	go test ./...

syntax:
	treerack generate -package-name ini -syntax ini/syntax.treerack > ini/syntax.go
	gofmt -w -s ./ini

fmt:
	gofmt -w -s .
	gofmt -w -s ./ini
