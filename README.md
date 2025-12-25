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

Returns subscriber info or an error payload.

Example response:

```json
{
  "msisdn": 380671234567,
  "billing_type": 1,
  "language_type": 0,
  "operator_type": 1,
  "change_date": "2024-05-09T09:25:16.482581Z"
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

### `GET /`

Health check, returns `200 OK`.

### `GET /metrics`

Plain text metric:

```
subscribers_total <count>
```

## Configuration

Environment variables are parsed via `github.com/caarlos0/env/v11`.

| Variable | Default | Description |
| --- | --- | --- |
| `SIS_DB_URL` | `postgresql://sis:EYTPu727BM2x3GY@localhost:5432/sis` | PostgreSQL DSN |
| `SIS_ADDR` | `:9001` | HTTP listen address |
| `SIS_CC` | `380` | Country code |
| `SIS_NDCS` | `67` | Comma-separated NDC list |
| `SIS_NDC_CAPACITY` | `10000000` | Capacity per NDC |
| `SIS_MSISDN_LENGTH` | `12` | Expected MSISDN length |
| `TZ` | `Europe/Prague` | Time zone (Docker) |

The `.env` in this repo overrides defaults for local Docker usage.

## Database

Schema is defined in `db/conf/init.sql`:

```sql
create schema sis;
create table sis.info (
  msisdn        bigint primary key,
  billing_type  smallint,
  language_type smallint,
  operator_type smallint,
  change_date   timestamp default now()
);
```

Note: Queries in `pkg/sis/store_pg.go` access `info` without schema.
Ensure `search_path` includes `sis` or update queries to `sis.info`.

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
./wrkb sis http://127.0.0.1:8080/subscribers/380671234567                                                                                               [23:05:20]

Process "sis" starts with:
cpu: 0.007197
threads: 6
mem: 9.6 MB
disk: 14 MB

┌────┬───────┬─────────┬─────┬────┬───────┐
│conn│    rps│  latency│  cpu│ thr│    mem│
├────┼───────┼─────────┼─────┼────┼───────┤
│   1│  19120│  98.19µs│ 0.77│  11│  18 MB│
│   2│  32759│ 117.85µs│ 1.15│  11│  19 MB│
│   3│  47480│  64.97µs│ 1.49│  12│  20 MB│
│   4│  52920│  77.89µs│ 1.63│  13│  20 MB│
│   5│  55940│  92.03µs│ 1.70│  13│  20 MB│
│   6│  54310│  115.2µs│ 1.99│  15│  21 MB│
│   7│  54400│ 130.53µs│ 2.22│  15│  21 MB│
│   8│  53960│ 149.34µs│ 2.66│  16│  21 MB│
│   9│  55820│ 164.38µs│ 1.94│  16│  21 MB│
│  10│  55970│ 180.49µs│ 2.72│  16│  21 MB│
│  12│  55790│ 220.16µs│ 2.43│  16│  21 MB│
│  16│  54960│ 288.89µs│ 2.35│  16│  22 MB│
│  32│  54740│ 583.87µs│ 2.20│  16│  22 MB│
│  64│  55050│   1.16ms│ 3.45│  16│  22 MB│
│ 128│  54760│   2.34ms│ 3.04│  17│  26 MB│
└────┴───────┴─────────┴─────┴────┴───────┘

Best: 5, rps: 55940, latency: 92.03µs
```
