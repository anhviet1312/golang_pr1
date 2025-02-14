package models

import "github.com/uptrace/bun"

type Story struct {
	bun.BaseModel `bun:"table:story"`
	ID            int64  `bun:"id,pk" json:"id"`
	Slug          string `bun:"slug" json:"slug"`
	Tittle        string `bun:"tiltle" json:"tile"`
	Description   string `bun:"description" json:"description"`
	Creator       string `bun:"creator" json:"creator"`
	CreatedAt     int64  `bun:"create_at" json:"created_at"`
	UpdatedAt     int64  `bun:"update_at" json:"updated_at"`
	Status        string `bun:"status" json:"status"`
	Image         string `bun:"image" json:"image"`
}

type Stories struct{}
