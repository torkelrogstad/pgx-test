package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

func realMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	if err != nil {
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)
	for i := 0; i < 10_000; i++ {
		group.Go(func() error {
			return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `SELECT pg_sleep(1)`)
				return err
			})
		})
	}

	time.AfterFunc(time.Second*15, func() {
		select {
		// already finished, don't bother dumping
		case <-ctx.Done():
			return

		default:
		}
		fmt.Println("timed out, dumping stack")
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)

		cancel()
	})

	return group.Wait()

}

func main() {
	err := realMain()
	if err != nil {
		log.Fatal(err)
	}
}
