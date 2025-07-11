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

-- name: GetCharactersByCoordinates :many
SELECT * from characters
WHERE position_x = $1 AND position_y = $2;

-- name: GetCharacterById :one
SELECT * from characters
WHERE id = $1;

-- name: GetCharacterByName :one
SELECT * from characters
where name = $1;

-- name: UpdateCharacterById :one
UPDATE characters
SET action_id = $1, 
	updated_at = NOW()
WHERE id = $2
RETURNING *;
