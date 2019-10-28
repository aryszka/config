SOURCES = $(shell find . -name "*.go")

.PHONY: cover showcover

default: build

deps:
	go get golang.org/x/tools/cmd/goimports
	go get github.com/aryszka/treerack/...
	go get ./...

build: $(SOURCES)
	go build

check: ini/syntax/syntax.treerack $(SOURCES)
	treerack check-syntax ini/syntax/syntax.treerack
	treerack check -syntax ini/syntax/syntax.treerack examples/skipper.ini
	go test ./...

.coverprofile: $(SOURCES)
	go test -coverprofile .coverprofile ./...

cover: .coverprofile
	go tool cover -func .coverprofile | grep -v syntax[.]go

showcover: .coverprofile
	go tool cover -html .coverprofile

syntax: ini/syntax/syntax.treerack
	treerack generate -package-name syntax -export -syntax ini/syntax/syntax.treerack > ini/syntax/syntax.go
	gofmt -w -s ./ini

fmt: $(SOURCES)
	gofmt -w -s .
	gofmt -w -s ./ini
	gofmt -w -s ./keys

precommit: fmt check

imports:
	goimports -w $(SOURCES)
