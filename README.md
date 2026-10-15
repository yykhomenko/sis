# SIS

Subscriber Information System — a small HTTP service that returns subscriber
profile data by MSISDN, backed by PostgreSQL (or an in-memory store for tests).

## Features

- HTTP API with validation for CC, NDC, and MSISDN length.
- PostgreSQL store with upsert support.
- In-memory store for benchmarks/tests.
- Minimal metrics endpoint.
- Docker Compose for local development.

## API

Base URL depends on `SIS_ADDR` (defaults to `:9001` in config; `.env` overrides for Docker).

### `GET /subscribers/:msisdn`

Returns subscriber or an error payload.

Example response:

```json
{
  "msisdn": 380671234567,
  "billing_type": 1,
  "language_type": 0,
  "operator_type": 1,
  "updated_at": "2024-05-09T09:25:16.482581Z"
}
```

Error payload:

```json
{
  "error_id": 1,
  "error_msg": "Not found"
}
```

Error IDs:

- `1` Not found
- `2` Invalid MSISDN format
- `3` Unsupported CC
- `4` Unsupported NDC

### `PUT /subscribers/:msisdn`

Upserts subscriber profile data.

Request body:

```json
{
  "billing_type": 2,
  "language_type": 1,
  "operator_type": 0
}
```

Response payload mirrors `GET /subscribers/:msisdn`.

### `GET /`

Health check, returns `200 OK`.

### `GET /metrics`

Plain text metric:

```
subscribers_total <count>
```

## Configuration

Environment variables are parsed via `github.com/caarlos0/env/v11`.

| Variable | Default                                    | Description |
| --- |--------------------------------------------| --- |
| `SIS_DB_URL` | `postgresql://sis:XXXX@localhost:5432/sis` | PostgreSQL DSN |
| `SIS_ADDR` | `:9001`                                    | HTTP listen address |
| `SIS_CC` | `380`                                      | Country code |
| `SIS_NDCS` | `67`                                       | Comma-separated NDC list |
| `SIS_NDC_CAPACITY` | `10000000`                                 | Capacity per NDC |
| `SIS_MSISDN_LENGTH` | `12`                                       | Expected MSISDN length |
| `TZ` | `Europe/Prague`                            | Time zone (Docker) |

The `.env` in this repo overrides defaults for local Docker usage.

## Database

Migrations are managed with Goose in `sql/schema`. Use:

```bash
make migration_up
make migration_down
```

Docker Compose uses `db/conf/init.sql` on first init; keep it aligned with migrations.

SQLC config lives in `sqlc.yaml` with schema in `sql/schema` and queries in `sql/queries`. Generated code goes to `internal/database`. Generate code via:

```bash
make generate_database_code
```

## Local Development

### Docker Compose

```bash
make start
```

Service ports:

- PostgreSQL: `localhost:5443`
- SIS API: `localhost:9001` (mapped to container port)

### Run Locally

```bash
go run ./cmd/sis
```

Example request:

```bash
curl http://localhost:9001/subscribers/380670000001
```

## Tests and Benchmarks

```bash
make test
make bench
```

## Build

```bash
make image
```

## Project Layout

- `cmd/sis` main entry point
- `internal/sis` core logic, server, and stores
- `internal/database` SQLC-generated DB layer
- `db/conf` database initialization
- `docker-compose.yml` local stack

## Bench
```
~/src/go/wrkb/wrkb -p=main http://127.0.0.1:9001/subscribers/__RANDI64_380670000001_380679999999__

⚙️  Preparing benchmark: 'main' [GET] for http://127.0.0.1:9001/subscribers/__RANDI64_380670000001_380679999999__
   Connections: [1 2 4 8 16 32 64 128 256] | Duration: 1s | Verbose: false

⚙️  Process: main
   CPU: 0.02s | Threads: 5 | Mem: 10 MB | Disk: 10 MB


┌────┬────────┬────────────┬────────┬────────┬────────┬────────┬─────┬────┬────────┐
│conn│     rps│     latency│    good│     bad│     err│    body│  cpu│ thr│     mem│
├────┼────────┼────────────┼────────┼────────┼────────┼────────┼─────┼────┼────────┤
│   1│   14210│     70.32µs│   14210│       0│       0│  1.7 MB│ 0.78│  10│   19 MB│
│   2│   21240│    94.084µs│   21240│       0│       0│  2.5 MB│ 1.43│  12│   20 MB│
│   4│   31801│   125.711µs│   31801│       0│       0│  3.8 MB│ 2.21│  14│   21 MB│
│   8│   40968│   195.208µs│   40968│       0│       0│  4.9 MB│ 3.18│  14│   22 MB│
│  16│   40727│   392.761µs│   40727│       0│       0│  4.8 MB│ 4.13│  18│   24 MB│
│  32│   50852│   629.262µs│   50852│       0│       0│  6.0 MB│ 4.54│  21│   28 MB│
│  64│   51741│  1.237009ms│   51741│       0│       0│  6.2 MB│ 4.56│  35│   36 MB│
│ 128│   52610│  2.433627ms│   52610│       0│       0│  6.3 MB│ 4.63│  35│   40 MB│
│ 256│   51890│  4.938424ms│   51890│       0│       0│  6.2 MB│ 4.63│  35│   49 MB│
└────┴────────┴────────────┴────────┴────────┴────────┴────────┴─────┴────┴────────┘

🏅  Best result: 32 connections | 50852 RPS | 629.262µs latency

```
