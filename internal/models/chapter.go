package models

import "github.com/uptrace/bun"

type Chapter struct {
	bun.BaseModel   `bun:"table:chapter"`
	ID              int64  `bun:"id,pk,autoincrement" json:"id"`
	Title           string `bun:"title" json:"title"`
	PreviousChapter int64  `bun:"previous_chapter" json:"previous_chapter"`
	AfterChapter    int64  `bun:"after_chapter" json:"after_chapter"`
	Publisher       string `bun:"publisher" json:"publisher"`
}
