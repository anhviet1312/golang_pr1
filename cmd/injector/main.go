package injector

import (
	"database/sql"
	"demo-cosebase/internal/services"
	"demo-cosebase/pkg/caching"
	"fmt"
	"github.com/hiendaovinh/toolkit/pkg/db"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"os"
)

func init() {
	godotenv.Load("../../.env") // for develop
	godotenv.Load("./.env")     // for production
}

func NewContainer(vs map[string]string) *do.Injector {
	injector := do.New()
	vs["API_MODE"] = os.Getenv("API_MODE")
	vs["API_ORIGINS"] = os.Getenv("API_ORIGINS")
	if vs["API_MODE"] == "" {
		vs["API_MODE"] = "production"
	}
	if vs["API_ORIGINS"] == "" {
		vs["API_ORIGINS"] = "*"
	}

	do.ProvideNamedValue(injector, "envs", vs)

	do.Provide(injector, func(injector *do.Injector) (*bun.DB, error) {
		connection := pgdriver.NewConnector(
			pgdriver.WithDSN(os.Getenv("DB_DSN")),
		)
		sqlDb := sql.OpenDB(connection)
		sqlDb.SetMaxOpenConns(50)
		sqlDb.SetMaxIdleConns(20)
		if err := sqlDb.Ping(); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		db := bun.NewDB(sqlDb, pgdialect.New())
		return db, nil
	})

	do.ProvideNamed(injector, "redis-db", func(i *do.Injector) (redis.UniversalClient, error) {
		clusterCacheRedisURL := os.Getenv("CLUSTER_REDIS_QUESTIONNAIRE")
		if clusterCacheRedisURL != "" {
			clusterOpts, err := redis.ParseClusterURL(clusterCacheRedisURL)
			if err != nil {
				return nil, err
			}
			return redis.NewClusterClient(clusterOpts), nil
		}
		return db.InitRedis(&db.RedisConfig{
			URL: os.Getenv("REDIS_QUESTIONNAIRE"),
		})
	})

	do.ProvideNamed(injector, "redis-cache", func(i *do.Injector) (redis.UniversalClient, error) {
		clusterCacheRedisURL := os.Getenv("CLUSTER_REDIS_CACHE")
		if clusterCacheRedisURL != "" {
			clusterOpts, err := redis.ParseClusterURL(clusterCacheRedisURL)
			if err != nil {
				return nil, err
			}
			return redis.NewClusterClient(clusterOpts), nil
		}
		return db.InitRedis(&db.RedisConfig{
			URL: os.Getenv("REDIS_CACHE"),
		})
	})

	do.ProvideNamed(injector, "redis-cache-readonly", func(i *do.Injector) (redis.UniversalClient, error) {
		var clusterOpts *redis.ClusterOptions
		var err error
		clusterCacheRedisReadOnlyURL := os.Getenv("CLUSTER_REDIS_CACHE_READONLY")
		if clusterCacheRedisReadOnlyURL != "" {
			clusterOpts, err = redis.ParseClusterURL(clusterCacheRedisReadOnlyURL)
		} else {
			clusterCacheRedisURL := os.Getenv("CLUSTER_REDIS_CACHE")
			if clusterCacheRedisURL != "" {
				clusterOpts, err = redis.ParseClusterURL(clusterCacheRedisURL)
			}
		}

		if err != nil {
			return nil, err
		}
		if clusterOpts != nil {
			clusterOpts.ReadOnly = true
			return redis.NewClusterClient(clusterOpts), nil
		}

		return db.InitRedis(&db.RedisConfig{
			URL: os.Getenv("REDIS_CACHE_READONLY"),
		})
	})

	do.Provide(injector, func(i *do.Injector) (caching.Cache, error) {
		dbRedis, err := do.InvokeNamed[redis.UniversalClient](i, "redis-cache")
		if err != nil {
			return nil, err
		}

		return caching.NewCacheRedis(dbRedis, false)
	})

	do.Provide(injector, func(i *do.Injector) (caching.ReadOnlyCache, error) {
		dbRedis, err := do.InvokeNamed[redis.UniversalClient](i, "redis-cache-readonly")
		if err != nil {
			return nil, err
		}

		return caching.NewCacheRedis(dbRedis, false)
	})

	do.Provide(injector, func(i *do.Injector) (*services.ServiceUser, error) {
		return services.NewServiceUser(injector)
	})

	return injector
}
