package main

import (
	"context"
	"database/sql"
	"demo-cosebase/internal/datastore"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func init() {
	godotenv.Load("../../.env") // for develop
	godotenv.Load("./.env")     // for production
}

func main() {
	app := &cli.App{
		Name: "migrate",
		Commands: []*cli.Command{
			commandMigration(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func commandMigration() *cli.Command {
	return &cli.Command{
		Name: "migrate",
		Action: func(c *cli.Context) error {
			ctx := context.Background()
			db, err := getDb()
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Start migrate user table")
			err = datastore.CreateTableUser(ctx, db)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Migration success")

			return nil
		},
	}
}

func getDb() (*bun.DB, error) {
	fmt.Println(os.Getenv("DB_DSN"))
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(os.Getenv("DB_DSN")),
	))

	db := bun.NewDB(sqldb, pgdialect.New())
	err := db.Ping()
	if err != nil {
		return nil, fmt.Errorf("loi ket noi db: %w", err)
	}
	return db, nil
}
