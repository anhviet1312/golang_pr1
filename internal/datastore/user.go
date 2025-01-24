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

func FindUserByEmail(ctx context.Context, db *bun.DB, email string) (*models.User, error) {
	user := &models.User{}
	err := db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func FindUserByID(ctx context.Context, db *bun.DB, ID int64) (*models.User, error) {
	user := &models.User{}
	err := db.NewSelect().Model(user).Where("id = ?", ID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func FindUserByEmailOrUsername(ctx context.Context, db *bun.DB, email, username string) (*models.User, error) {
	user := &models.User{}
	err := db.NewSelect().Model(user).Where("email = ?", email).WhereOr("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateUser(ctx context.Context, db *bun.DB, user *models.User) (*models.User, error) {
	_, err := db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUser(ctx context.Context, db *bun.DB, user *models.User) (*models.User, error) {
	_, err := db.NewUpdate().Model(user).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}
