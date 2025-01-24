package models

type Story struct {
	ID          int64  `bun:"id,pk" json:"id"`
	Tittle      string `bun:"tiltle" json:"tile"`
	Description string `bun:"description" json:"description"`
	Creator     string `bun:"creator" json:"creator"`
	Image       string `bun:"image" json:"image"`
}
