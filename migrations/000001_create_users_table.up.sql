CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password BYTEA NOT NULL,
    email TEXT UNIQUE,
    telegram TEXT UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT now()
)