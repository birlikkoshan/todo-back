// @title           Todo API
// @version         1.0
// @description     Todo API with auth, search, overdue.
// @host            localhost:8080
// @BasePath        /api/v1
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Worker/internal/app"
	"Worker/internal/config"

	_ "Worker/docs"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	application, err := app.New(cfg)
	if err != nil {
		panic(err)
	}
	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      application.Router(),
		ReadTimeout:  cfg.HTTP.ReadTimeout.Duration(),
		WriteTimeout: cfg.HTTP.WriteTimeout.Duration(),
		IdleTimeout:  cfg.HTTP.IdleTimeout.Duration(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}

	if err := application.Close(ctx); err != nil {
		panic(err)
	}
}
