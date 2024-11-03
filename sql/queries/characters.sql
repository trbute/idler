-- name: CreateCharacter :one
INSERT INTO characters(id, user_id, name, created_at, updated_at)
VALUES (
	gen_random_uuid(),
	$1,
	$2,
	NOW(),
	NOW()
)
RETURNING *;

-- name: GetCharacterById :one
SELECT * from characters
WHERE id = $1;

