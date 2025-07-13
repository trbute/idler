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

-- name: UpdateCharacterByIdWithTargetAndAmount :one
UPDATE characters
SET action_id = $1,
	action_target = $2,
	action_amount_limit = $3,
	action_amount_progress = 0,
	updated_at = NOW()
WHERE id = $4
RETURNING *;

-- name: UpdateCharacterProgress :one
UPDATE characters
SET action_amount_progress = $1,
	updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: SetCharacterToIdleAndResetGathering :one
UPDATE characters
SET action_id = $1,
	action_target = NULL,
	action_amount_limit = NULL,
	action_amount_progress = 0,
	updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: GetActiveCharacters :many
SELECT * FROM characters
WHERE action_id != 1;
