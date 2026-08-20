default_os := `go env GOOS`
default_arch := `go env GOARCH`

default:
    @just --list

fmt:
    gofmt -w $(go list -f '{{"{{"}}.Dir{{"}}"}}' ./...)

lint:
    golangci-lint run

generate:
    go generate ./...

verify-generate:
    just generate
    git diff --exit-code

test:
    go test -v -race -cover ./...

build os=default_os arch=default_arch:
    CGO_ENABLED=0 GOOS={{os}} GOARCH={{arch}} go build -o bin/gopherquest-{{os}}-{{arch}} ./cmd/gopherquest

run:
    go run ./cmd/gopherquest

clean:
    rm -rf bin/
    go clean
