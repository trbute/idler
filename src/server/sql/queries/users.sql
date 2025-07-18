-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password, surname)
VALUES (
	gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2,
	$3
)
RETURNING *;

-- name: GetUserById :one
SELECT * from users
WHERE id = $1;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: UpdateUserById :one
UPDATE users
SET email = $2, 
	hashed_password = $3,
	updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetSurnameById :one
SELECT surname FROM users
WHERE id = $1;
