package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/maks3201/sns-alert-service/config"
	"github.com/maks3201/sns-alert-service/internal/alertmanager"
	"github.com/maks3201/sns-alert-service/internal/aws"
	health "github.com/maks3201/sns-alert-service/internal/status"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	configFilePath := flag.String("config", "config/config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg := config.LoadConfig(*configFilePath)

	awsClient, err := aws.InitSNSClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize AWS client: %v", err)
	}

	alertHandler := alertmanager.NewHandler(cfg, awsClient)

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		health.HealthHandler(w, r, awsClient)
	})

	http.HandleFunc("/alert", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if log.IsLevelEnabled(log.DebugLevel) {
			log.Debugf("Received alert: %s", string(bodyBytes))
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		alertHandler.SNSHandler(w, r)
	})

	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              ":8080",
		ReadTimeout:       time.Duration(cfg.Timeouts.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout:      time.Duration(cfg.Timeouts.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:       time.Duration(cfg.Timeouts.Server.IdleTimeoutSeconds) * time.Second,
		ReadHeaderTimeout: time.Duration(cfg.Timeouts.Server.ReadHeaderTimeoutSeconds) * time.Second,
	}

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

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	cancel()
	wg.Wait()

	log.Info("Server exiting")
}
