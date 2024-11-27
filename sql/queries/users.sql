-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2
)
RETURNING *;

-- name: WipeUsers :execresult
DELETE FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $2,
	hashed_password = $3,
	updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UnlockRed :exec
UPDATE users
SET chirpy_red = true
WHERE id = $1;