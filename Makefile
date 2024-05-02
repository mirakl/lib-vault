
goimports:
	go get -v golang.org/x/tools/cmd/goimports

fmt: goimports
	goimports -w .

test:
	go test ./... -v

upgrade:
	go install github.com/marwan-at-work/mod/cmd/mod@latest
	mod upgrade

check: setup
	bin/golangci-lint run --timeout=600s

setup: ./bin/golangci-lint

./bin/golangci-lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.57.2
