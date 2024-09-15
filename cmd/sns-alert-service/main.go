package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maks3201/sns-alert-service/config"
	"github.com/maks3201/sns-alert-service/internal/alertmanager"
	"github.com/maks3201/sns-alert-service/internal/aws"
	"github.com/maks3201/sns-alert-service/internal/health"
	"github.com/maks3201/sns-alert-service/internal/ready"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	cfg := config.LoadConfig()

	if err := aws.InitSNSClient(cfg); err != nil {
		log.Fatalf("Failed to initialize AWS client: %v", err)
	}

	http.HandleFunc("/healthz", health.HealthHandler)
	http.HandleFunc("/alert", alertmanager.SNSHandler)
	http.HandleFunc("/ready", ready.ReadyHandler)

	server := &http.Server{Addr: ":80"}

	go func() {
		log.Infof("Server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exiting")
}
