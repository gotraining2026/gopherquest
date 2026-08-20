CREATE TABLE IF NOT EXISTS campaigns (
    id            BIGSERIAL PRIMARY KEY,
    player_name   TEXT NOT NULL CHECK (char_length(btrim(player_name)) BETWEEN 1 AND 40),
    title         TEXT NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 80),
    module_path   TEXT NOT NULL CHECK (char_length(module_path) BETWEEN 3 AND 200),
    difficulty    TEXT NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
    stage         TEXT NOT NULL DEFAULT 'module' CHECK (stage IN ('module', 'errors', 'channel', 'complete')),
    score         INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
