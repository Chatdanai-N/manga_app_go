package main

import (
	"context"
	"log/slog"
	"manga_app/config"
	"manga_app/database"
	"manga_app/middleware"
	"manga_app/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	middleware.InitLogger()


	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config ", "error", err)
		os.Exit(1)
	}

	database.DB, err = database.ConnectDatabase(cfg)
	if err != nil {
		slog.Error("database connect error", "error", err)
		os.Exit(1)
	}

	// Setup Router & Start Server
	r := router.SetupRouter(database.DB)

	// Setup Router & HTTP Server Structure
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start Server ใน Goroutine เพื่อไม่ให้ Block Graceful Shutdown Signal
	go func() {
		slog.Info("server is starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server is stopping", "error", err)
			os.Exit(1)
		}
	}()

	//Graceful Shutdown Listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server", "port", cfg.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server force to shutdown", "error", err)
	}
	slog.Info("server exited cleanly")
}
