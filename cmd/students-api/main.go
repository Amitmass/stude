package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Amitmass/students-api/internal/config"
	"github.com/Amitmass/students-api/internal/http/handlers/student"
	"github.com/Amitmass/students-api/internal/storage/sqlite"
)

func main() {
	// Load Config

	cfg := config.MustLoad()

	// Database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storageinitialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))
	// Setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(storage))

	// Setup Server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	slog.Info("server started", slog.String("address", cfg.Addr))
	// fmt.Printf("server started %s", cfg.HTTPServer.Addr)

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start server")
		}
	}()

	<-done

	slog.Info("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		slog.Error("failed to shutdown", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown successfully.")

}
