package main

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target internal/api --package api --clean api/openapi/gopherquest.openapi.yaml
//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
