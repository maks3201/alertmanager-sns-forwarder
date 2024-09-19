package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	awsClient, err := aws.InitSNSClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize AWS client: %v", err)
	}

	alertHandler := alertmanager.NewHandler(cfg, awsClient)

	http.HandleFunc("/healthz", health.HealthHandler)
	http.HandleFunc("/alert", alertHandler.SNSHandler)
	http.HandleFunc("/ready", ready.NewHandler(awsClient).ReadyHandler)

	server := &http.Server{Addr: ":80"}

	// Start batching goroutine
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		alertHandler.ProcessBatches(ctx)
	}()

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
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Stop batching goroutine
	cancel()
	wg.Wait()

	log.Info("Server exiting")
}
