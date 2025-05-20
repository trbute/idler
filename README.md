# idler

## Go api-based idler rpg

### ENV VARS

    create .env (copy & modify .env.example)

## Running via Docker

### create db & redis, run migrations, run server

    docker compose up --build

### delete containers & volumes

    docker compose down -v