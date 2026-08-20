-- name: CreateCampaign :one
INSERT INTO campaigns (player_name, title, module_path, difficulty)
VALUES ($1, $2, $3, $4)
RETURNING id, player_name, title, module_path, difficulty, stage, score, created_at, updated_at;

-- name: ListCampaigns :many
SELECT id, player_name, title, module_path, difficulty, stage, score, created_at, updated_at
FROM campaigns
ORDER BY id;

-- name: GetCampaign :one
SELECT id, player_name, title, module_path, difficulty, stage, score, created_at, updated_at
FROM campaigns
WHERE id = $1;

-- name: UpdateCampaign :one
UPDATE campaigns
SET player_name = $2,
    title = $3,
    module_path = $4,
    difficulty = $5,
    updated_at = now()
WHERE id = $1
  AND stage = 'module'
RETURNING id, player_name, title, module_path, difficulty, stage, score, created_at, updated_at;

-- name: DeleteCampaign :execrows
DELETE FROM campaigns
WHERE id = $1;

-- name: AdvanceCampaign :one
UPDATE campaigns
SET stage = $2,
    score = score + $3,
    updated_at = now()
WHERE id = $1
  AND stage = $4
RETURNING id, player_name, title, module_path, difficulty, stage, score, created_at, updated_at;
