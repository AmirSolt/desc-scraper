-- Amirali Soltani

CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    yt_id TEXT NOT NULL UNIQUE,
    thumbnail_url TEXT NOT NULL,
    handle TEXT NOT NULL,
    title TEXT NOT NULL
);
CREATE UNIQUE INDEX yt_ch_idx ON channels ("yt_id");


CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    yt_id TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,

    channel_id INT NOT NULL REFERENCES channels(id)
);

CREATE UNIQUE INDEX yt_vid_idx ON videos ("yt_id");
