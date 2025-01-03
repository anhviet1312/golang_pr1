package datastore

import (
	"context"
	"demo-cosebase/internal/models"
	"github.com/uptrace/bun"
)

func CreateTableUser(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().Model((*models.User)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func FindUserByUsername(ctx context.Context, db *bun.DB, username string) (*models.User, error) {
	user := &models.User{}
	err := db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}
