# idler

## Go api-based idler rpg

### Migrations

Use Postgres + [Goose](https://github.com/pressly/goose) 

From `sql/schema` directory:

    goose postgres {{databaseConnectionURL}} up

### Query builder

Use [SQLC](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html)

### ENV VARS

    DB_URL = "postgres://{{user}}:{{password}}@localhost:5432/idler?sslmode=disable"
    PLATFORM = "dev"
    JWT_SECRET = "{{jwt}}"
    
Generate JWT secret token: 

    openssl rand -base64 64
