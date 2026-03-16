package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code-server/internal/config"
	"code-server/internal/server"
)

// 打包时通过 -ldflags "-X main.buildEnv=prod" 注入，未设置时为空
var buildEnv string

func appEnv() string {
	if v := os.Getenv("APP_ENV"); v != "" {
		return v
	}
	if buildEnv != "" {
		return buildEnv
	}
	return "development"
}

func envOrUnknown(s string) string {
	if s == "" {
		return "development"
	}
	return s
}

func main() {
	config.LoadEnv() // 加载 .env / .env.{APP_ENV} / .env.local（类似 Node dotenv）

	port := os.Getenv("PORT")
	if port == "" {
		switch appEnv() {
		case "prod":
			port = "8090"
		case "development", "":
			port = "8080"
		default:
			port = "8080"
		}
	}

	srv := server.New(port)

	go func() {
		log.Printf("Server listening on :%s (APP_ENV=%s)", port, envOrUnknown(appEnv()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
