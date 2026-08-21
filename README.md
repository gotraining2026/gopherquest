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
- [x] Cobra campaign commands: `add`, `new`, `list`, `get`, `answer`, and `delete`
- [x] Bubble Tea interactive campaign screen
- [x] Structured request logging
- [x] Runtime configuration
- [x] API container and Compose startup
- [ ] Coverage closure, final checks, and demo documentation

## Run the Demo

Start the API and PostgreSQL:

```text
docker compose up --build -d
```

Create a campaign and use the ID printed by the command:

```text
go run ./cmd/gopherquestctl campaigns add --player "Demo Player" --title "Protocol Run" --module "github.com/spf13/cobra" --difficulty beginner
# Replace 1 with the ID printed by the add command.
go run ./cmd/gopherquestctl campaigns list
go run ./cmd/gopherquestctl campaigns get --id 1
go run ./cmd/gopherquestctl campaigns play
```

The CLI provides `add`/`new`, `list`, `get`, `answer`, and `delete` commands.
The TUI can list and select campaigns, create a new campaign with `N`, delete
the selected campaign with `D`, refresh with `R`, and play challenge answers.
Passing `--id 1` skips the campaign list and opens that campaign directly.

In the new-campaign form, the title field shows an example placeholder. The
module and difficulty fields offer Left/Right preset selection while the
player, title, and module fields also support manual entry.

Delete one campaign with:

```text
go run ./cmd/gopherquestctl campaigns delete --id 1
```

The scripted answer flow is also available:

```text
go run ./cmd/gopherquestctl campaigns answer --id 1 --answer v1
go run ./cmd/gopherquestctl campaigns answer --id 1 --answer errors.Is
go run ./cmd/gopherquestctl campaigns answer --id 1 --answer channel
```

All difficulty levels use the same challenge answers. Difficulty changes the
score multiplier:

| Difficulty | Answers | Final score |
| --- | --- | ---: |
| `beginner` | `v1`, `errors.Is`, `channel` | 400 |
| `intermediate` | `v1`, `errors.Is`, `channel` | 800 |
| `advanced` | `v1`, `errors.Is`, `channel` | 1,200 |

Create an intermediate or advanced campaign by changing the difficulty flag:

```text
go run ./cmd/gopherquestctl campaigns new --player "Demo Player" --title "Intermediate Run" --module "github.com/spf13/cobra" --difficulty intermediate
go run ./cmd/gopherquestctl campaigns new --player "Demo Player" --title "Advanced Run" --module "github.com/spf13/cobra" --difficulty advanced
```

## Challenge Content

Challenge content is currently hand-written in the domain package. There is no
question configuration file:

- `internal/campaign/campaign.go` defines each stage prompt and hint.
- `internal/campaign/challenge.go` evaluates answers, transitions stages, and
  calculates points.
- `internal/campaign/challenge_test.go` covers answer and scoring rules.

The first answer is derived from live `moduleInfo.latestVersion`; the other
answers are the normalized values `errors.Is` and `channel`. `.env` settings
configure the service and provider, not the questions or answers.

Stop the local services afterward:

```text
docker compose down
```

To reset all local campaign data, use `docker compose down -v` before starting
the services again. Database-wide reset is intentionally a Docker operation,
not a CLI or API command.
