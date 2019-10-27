.PHONY: .coverprofile

default: build

deps:
	go get github.com/aryszka/treerack/...
	go get ./...

build:
	go build

check:
	treerack check-syntax ini/syntax/syntax.treerack
	treerack check -syntax ini/syntax/syntax.treerack examples/skipper.ini
	go test ./...

.coverprofile:
	go test -coverprofile .coverprofile ./...

cover: .coverprofile
	go tool cover -func .coverprofile | grep -v syntax[.]go

showcover: .coverprofile
	go tool cover -html .coverprofile

syntax:
	treerack generate -package-name syntax -export -syntax ini/syntax/syntax.treerack > ini/syntax/syntax.go
	gofmt -w -s ./ini

fmt:
	gofmt -w -s .
	gofmt -w -s ./ini
	gofmt -w -s ./keys

precommit: fmt check
