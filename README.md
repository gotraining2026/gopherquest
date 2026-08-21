# GopherQuest

GopherQuest is a local, single-player Go learning game. The Cobra CLI talks to
a contract-first HTTP API, which persists campaigns in PostgreSQL and uses the
Go Module Proxy for live module information.

## Status

- [x] Architecture context diagram
- [x] Go module and pinned `ogen`/`sqlc` tooling
- [x] OpenAPI 3.0.3 contract
- [x] Generated `ogen` server and client
- [x] PostgreSQL schema, queries, Compose health check, and generated `sqlc` code
- [x] Campaign domain values, validation, errors, stages, and tests
- [x] Go Module Proxy client and fixture tests
- [x] PostgreSQL domain adapter and campaign service
- [x] CRUD HTTP handlers, routing, and health check
- [x] Challenge answer evaluation and progress updates
- [ ] Cobra campaign commands
- [ ] Structured request logging
- [x] Runtime configuration
- [x] API container and Compose startup
- [ ] Coverage closure, final checks, and demo documentation
