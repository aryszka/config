.PHONY: .coverprofile

default: build

deps:
	go get github.com/aryszka/treerack/...

build:
	go build

check:
	treerack check-syntax ini/syntax.treerack
	treerack check -syntax ini/syntax.treerack examples/skipper.ini
	go test ./...

.coverprofile:
	go test -coverprofile .coverprofile ./...

cover: .coverprofile
	go tool cover -func .coverprofile | grep -v syntax[.]go

showcover: .coverprofile
	go tool cover -html .coverprofile

syntax:
	treerack generate -package-name ini -syntax ini/syntax.treerack > ini/syntax.go
	gofmt -w -s ./ini

fmt:
	gofmt -w -s .
	gofmt -w -s ./ini

precommit: fmt check
