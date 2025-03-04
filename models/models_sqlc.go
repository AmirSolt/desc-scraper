// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Channel struct {
	ID           int32              `json:"id"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	YtID         string             `json:"yt_id"`
	ThumbnailUrl string             `json:"thumbnail_url"`
	Handle       string             `json:"handle"`
	Title        string             `json:"title"`
}

type Video struct {
	ID          int32              `json:"id"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	YtID        string             `json:"yt_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	ChannelID   int32              `json:"channel_id"`
}
