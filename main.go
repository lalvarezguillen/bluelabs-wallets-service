package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := initApp()

	go func() {
		if err := e.Start(os.Getenv("LISTEN_ON")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	// SIGINT to handle ctrl+C
	// SIGTERM to handle 'docker stop'
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	received := <-quit
	e.Logger.Errorf("received signal: %+v\n", received)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func initApp() *echo.Echo {
	db, err := NewDB(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
	if err != nil {
		panic(err)
	}

	wStore := NewWalletStore(db)
	wService := NewWalletService(wStore)

	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())

	wallets := e.Group("/wallets")
	wc := NewWalletController(wService)
	wc.Register(wallets)

	return e
}
