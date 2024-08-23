build:
	CGO_ENABLED=0 go build -o gctl github.com/jlewi/gctl

tidy-go:
	gofmt -s -w .
	goimports -w .

tidy: tidy-go

lint-go:
	# golangci-lint automatically searches up the root tree for configuration files.
	golangci-lint run

lint: lint-go

test:
	go test -v ./...