SHELL := /bin/bash

all: \
	commitlint \
	go-mod-tidy \
	go-lint \
	go-review \
	go-test \
	readme \
	git-verify-nodiff

include tools/commitlint/rules.mk
include tools/git-verify-nodiff/rules.mk
include tools/golangci-lint/rules.mk
include tools/goreview/rules.mk
include tools/semantic-release/rules.mk
include tools/snippet/rules.mk

.PHONY: go-test
go-test:
	$(info [$@] running Go tests...)
	@go test -count 1 -cover -race ./...

.PHONY: go-mod-tidy
go-mod-tidy:
	$(info [$@] tidying Go module files...)
	@go mod tidy -v

.PHONY: readme
readme: $(snippet)
	$(info [$@] inserting usage in README...)
	@go run go.einride.tech/cloudrunner/examples/cmd/grpc-server -help | \
		$(snippet) README.md "usage"
