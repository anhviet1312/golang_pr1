package main

import (
	"context"
	"demo-cosebase/cmd/injector"
	"demo-cosebase/internal/api/handler"
	"github.com/joho/godotenv"
	"github.com/samber/do"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func init() {
	godotenv.Load("../../.env") // for develop
	godotenv.Load("./.env")     // for production
}
func main() {
	vs := map[string]string{}
	container := injector.NewContainer(vs)
	app := &cli.App{
		Name: "api",
		Commands: []*cli.Command{
			commandServer(container),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func commandServer(container *do.Injector) *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "start the web server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Value: "0.0.0.0:8080",
				Usage: "serve address",
			},
		},
		Action: func(c *cli.Context) error {
			vs := do.MustInvokeNamed[map[string]string](container, "envs")
			router, err := handler.New(&handler.Config{
				Container: container,
				Mode:      vs["API_MODE"],
				Origins:   strings.Split(vs["API_ORIGINS"], ","),
			})
			if err != nil {
				return err
			}

			srv := &http.Server{
				Addr:    c.String("addr"),
				Handler: router,
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			errWg, errCtx := errgroup.WithContext(ctx)

			errWg.Go(func() error {
				log.Printf("ListenAndServe: %s\n", c.String("addr"))
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					return err
				}
				return nil
			})

			errWg.Go(func() error {
				<-errCtx.Done()
				return srv.Shutdown(context.TODO())
			})

			return errWg.Wait()
		},
	}
}
