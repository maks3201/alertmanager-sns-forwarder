package alertmanager

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
    "github.com/maks3201/sns-alert-service/config"
    "github.com/maks3201/sns-alert-service/internal/aws"
)

type AlertmanagerPayload struct {
    Receiver          string `json:"receiver"`
    Status            string `json:"status"`
    Alerts            []Alert `json:"alerts"`
    GroupLabels       map[string]string `json:"groupLabels"`
    CommonLabels      map[string]string `json:"commonLabels"`
    CommonAnnotations map[string]string `json:"commonAnnotations"`
    ExternalURL       string `json:"externalURL"`
    Version           string `json:"version"`
    GroupKey          string `json:"groupKey"`
}

type Alert struct {
    Status      string            `json:"status"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    StartsAt    string            `json:"startsAt"`
    EndsAt      string            `json:"endsAt"`
    GeneratorURL string           `json:"generatorURL"`
}

func SNSHandler(w http.ResponseWriter, r *http.Request) {
    cfg := config.LoadConfig()

    var payload AlertmanagerPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Failed to parse request body", http.StatusBadRequest)
        log.Printf("Error parsing request: %v", err)
        return
    }

    message := formatAlertMessage(payload)

    if err := aws.PublishToSNS(cfg.SNSTopicARN, message); err != nil {
        http.Error(w, fmt.Sprintf("Failed to send message to SNS: %v", err), http.StatusInternalServerError)
        log.Printf("Error sending message to SNS: %v", err)
        return
    }

    log.Printf("Alert sent to SNS topic: %s", cfg.SNSTopicARN)
    fmt.Fprintf(w, "Alert sent to SNS topic: %s", cfg.SNSTopicARN)
}

func formatAlertMessage(payload AlertmanagerPayload) string {
    var message strings.Builder

    for _, alert := range payload.Alerts {
        fmt.Fprintf(&message, "[%s] (%s %s %s)\n", strings.ToUpper(alert.Status), alert.Labels["alertname"], alert.Labels["instance"], alert.Labels["severity"])
        fmt.Fprintf(&message, "Starts at: %s\n", alert.StartsAt)
        if alert.EndsAt != "" {
            fmt.Fprintf(&message, "Ends at: %s\n", alert.EndsAt)
        }
        fmt.Fprintf(&message, "Labels:\n")
        for key, value := range alert.Labels {
            fmt.Fprintf(&message, "%s = %s\n", key, value)
        }
        fmt.Fprintf(&message, "Annotations:\n")
        for key, value := range alert.Annotations {
            fmt.Fprintf(&message, "%s = %s\n", key, value)
        }
        fmt.Fprintf(&message, "Generator URL: %s\n", alert.GeneratorURL)
        fmt.Fprintf(&message, "\n")
    }
    return message.String()
}
