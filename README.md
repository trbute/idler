# idler

## Go api-based idler rpg

### ENV VARS

    create .env (copy/modify .env.example)

## Running via Docker

### create db, run migrations, run server

    docker compose up --build

### delete containers & volumes

    docker compose down -v

## Running locally

### Migrations

Use Postgres + [Goose](https://github.com/pressly/goose)

From `sql/schema` directory:

    goose postgres {{databaseConnectionURL}} up

### Query builder

Use [SQLC](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html)

Generate JWT secret token:

    openssl rand -base64 64
