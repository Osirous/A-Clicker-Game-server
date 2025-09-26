-- +goose Up
CREATE TABLE savedata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
     -- BYTEA is PostgreSQL-specific; for MySQL use BLOB, for SQL Server use VARBINARY(MAX) Ended up switch this to TEXT, but left this comment here for my edification.
    savedata TEXT NOT NULL, 
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE savedata;