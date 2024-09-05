package main

import (
    "log"
    "net/http"
    "github.com/maks3201/sns-alert-service/config"
    "github.com/maks3201/sns-alert-service/internal/alertmanager"
    "github.com/maks3201/sns-alert-service/internal/health"
    "github.com/maks3201/sns-alert-service/internal/aws"
)


func main() {
    // Load configuration and environment variables
    cfg := config.LoadConfig()

    // Initialize AWS SNS client
    if err := aws.InitSNSClient(cfg); err != nil {
        log.Fatalf("Failed to initialize AWS client: %v", err)
    }

    // Define HTTP route and handler
    http.HandleFunc("/healthz", health.HealthHandler) 
    http.HandleFunc("/alert", alertmanager.SNSHandler)

    // Start HTTP server
    log.Printf("Server started on port 80")
    if err := http.ListenAndServe(":80", nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
