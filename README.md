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

Base URL depends on `SIS_ADDR` (defaults to `:9001` in config, `:8080` in `.env`).

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
- `10` Internal

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

SQLC config lives in `sqlc.yaml` with schema in `db/conf/init.sql` and queries in `sql/queries`. Generate code via:

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
- SIS API: `localhost:9001` (mapped to container `:8080`)

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
- `pkg/sis` core logic, server, and stores
- `db/conf` database initialization
- `docker-compose.yml` local stack

## Bench
```
~/src/go/wrkb/wrkb -p=main http://127.0.0.1:9001/subscribers/__RANDI64_380670000001_380670099999__
false
⚙️  Preparing benchmark: 'main' [GET] for http://127.0.0.1:9001/subscribers/__RANDI64_380670000001_380670099999__
   Connections: [1 2 4 8 16 32 64 128 256] | Duration: 1s | Verbose: false

⚙️  Process: main
   CPU: 0.02s | Threads: 6 | Mem: 10 MB | Disk: 10 MB


┌────┬────────┬────────────┬────────┬────────┬────────┬────────┬─────┬────┬────────┐
│conn│     rps│     latency│    good│     bad│     err│    body│  cpu│ thr│     mem│
├────┼────────┼────────────┼────────┼────────┼────────┼────────┼─────┼────┼────────┤
│   1│   13956│     71.56µs│   13956│       0│       0│  1.7 MB│ 0.78│  10│   19 MB│
│   2│   20970│    95.298µs│   20970│       0│       0│  2.5 MB│ 1.43│  11│   20 MB│
│   4│   31193│   128.163µs│   31193│       0│       0│  3.7 MB│ 2.26│  14│   21 MB│
│   8│   40859│   195.744µs│   40859│       0│       0│  4.9 MB│ 3.22│  14│   22 MB│
│  16│   44104│   362.789µs│   44104│       0│       0│  5.2 MB│ 4.35│  20│   24 MB│
│  32│   47556│   672.846µs│   47556│       0│       0│  5.7 MB│ 4.57│  25│   28 MB│
│  64│   54408│  1.176608ms│   54408│       0│       0│  6.5 MB│ 4.58│  43│   37 MB│
│ 128│   53761│  2.380874ms│   53761│       0│       0│  6.4 MB│ 4.68│  43│   41 MB│
│ 256│   48416│  5.294847ms│   48416│       0│       0│  5.8 MB│ 4.51│  43│   50 MB│
└────┴────────┴────────────┴────────┴────────┴────────┴────────┴─────┴────┴────────┘

💫  Best result: 64 connections | 54408 RPS | 1.176608ms latency

```
