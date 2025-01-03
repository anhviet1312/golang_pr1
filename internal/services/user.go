package services

import (
	"context"
	"database/sql"
	"demo-cosebase/internal/datastore"
	"demo-cosebase/internal/models"
	"demo-cosebase/pkg/caching"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

type ServiceUser struct {
	container          *do.Injector
	redisDB            redis.UniversalClient
	postgresDB         *bun.DB
	readonlyPostgresDB *bun.DB
	cache              caching.Cache
}

func NewServiceUser(container *do.Injector) (*ServiceUser, error) {
	db, err := do.InvokeNamed[redis.UniversalClient](container, "redis-db")
	if err != nil {
		return nil, err
	}

	postgresDB, err := do.Invoke[*bun.DB](container)
	if err != nil {
		return nil, err
	}

	cache, err := do.Invoke[caching.Cache](container)
	if err != nil {
		return nil, err
	}

	readonlyPostgresDB, err := do.InvokeNamed[*bun.DB](container, "db-readonly")
	if err != nil {
		return nil, err
	}

	return &ServiceUser{container, db, postgresDB, readonlyPostgresDB, cache}, nil
}

func (service *ServiceUser) Authenticate(ctx context.Context, username string, password string) (*models.User, error) {
	user, err := datastore.FindUserByUsername(ctx, service.postgresDB, username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if !checkPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
