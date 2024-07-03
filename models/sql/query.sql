

-- =========================================
-- events


-- name: SearchVideoDescs :many
SELECT *
FROM videos
WHERE desc_fts @@ to_tsquery($1);

-- name: CreateVideos :copyfrom
INSERT INTO videos (
    yt_id,
    title,
    tags,
    default_language,
    description,
    live_broadcast_content,
    published_at,
    channel_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);

