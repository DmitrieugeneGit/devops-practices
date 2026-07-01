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

	"tasks-app/internal/config"
	"tasks-app/internal/database"
	"tasks-app/internal/handlers"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("подключение к БД: %v", err)
	}
	defer db.Close()
	log.Println("подключение к PostgreSQL установлено")

	h := handlers.New(db)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      h.Routes(cfg.FrontendDir),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в отдельной горутине.
	go func() {
		log.Printf("HTTP-сервер слушает на %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ошибка сервера: %v", err)
		}
	}()

	// Ожидаем сигнал завершения.
	<-ctx.Done()
	log.Println("получен сигнал завершения, останавливаемся...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("ошибка graceful shutdown: %v", err)
		os.Exit(1)
	}
	log.Println("сервер остановлен")
}
