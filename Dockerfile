# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOEXPERIMENT=jsonv2 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-w -s" -o /out/gopherquest ./cmd/gopherquest

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S gopherquest \
    && adduser -S -G gopherquest gopherquest

COPY --from=builder /out/gopherquest /usr/local/bin/gopherquest

USER gopherquest
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --retries=10 CMD wget -q -O - http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/gopherquest"]
