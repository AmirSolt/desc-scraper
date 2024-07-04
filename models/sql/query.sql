

-- =========================================
-- events


-- name: SearchVideoDescs :many
SELECT *
FROM videos
WHERE desc_fts @@ to_tsquery($1);

-- name: CreateChannel :one
INSERT INTO channels (
    yt_id,
    thumbnail_url,
    handle,
    title
) VALUES ($1,$2,$3,$4) RETURNING *;

-- name: GetChannelByYTID :one
SELECT * FROM channels WHERE yt_id = $1 ;




-- name: CreateVideo :one
INSERT INTO videos (
    yt_id,
    title,
    description,
    published_at,
    channel_id
) VALUES ($1,$2,$3,$4,$5) RETURNING *;


-- name: GetVideoByYTID :one
SELECT * FROM videos WHERE yt_id = $1 ;
