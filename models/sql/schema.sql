-- Amirali Soltani

CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    yt_id TEXT NOT NULL UNIQUE
    -- thumbnail
    -- handle
    -- title
);
CREATE UNIQUE INDEX yt_id_idx ON channels ("yt_id");


CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    yt_id TEXT NOT NULL UNIQUE,

    title TEXT NOT NULL,
    tags TEXT NOT NULL,
    default_language TEXT NOT NULL,
    description TEXT NOT NULL,
    desc_fts tsvector GENERATED ALWAYS AS (
        to_tsvector('english', description)
    ) STORED,
    live_broadcast_content BOOLEAN NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,

    channel_id INT NOT NULL REFERENCES channels(id)
);

CREATE UNIQUE INDEX yt_id_idx ON videos ("yt_id");
create index videos_fts on videos using gin (desc_vectors);
