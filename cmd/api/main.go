// @title           Todo API
// @version         1.0
// @description     Todo API with auth, search, overdue.
// @host            localhost:8080
// @BasePath        /api/v1
package main

import (
	"context"
	"errors"
	"log"
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
		log.Fatalf("config: %v", err)
	}
	log.Printf("config loaded, connecting to DB and Redis...")

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("app init: %v", err)
	}
	log.Printf("app ready, starting HTTP server")
	server := &http.Server{
		Addr:         "0.0.0.0:" + cfg.HTTP.Port,
		Handler:      application.Router(),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		log.Printf("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
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
