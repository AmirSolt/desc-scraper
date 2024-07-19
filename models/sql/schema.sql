-- Amirali Soltani

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    yt_id TEXT NOT NULL UNIQUE,
    thumbnail_url TEXT NOT NULL,
    handle TEXT NOT NULL,
    title TEXT NOT NULL
);


CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    yt_id TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,

    channel_id INT NOT NULL REFERENCES channels(id)
);


CREATE INDEX videos_description_trgm_idx ON videos USING GIN (description gin_trgm_ops);
