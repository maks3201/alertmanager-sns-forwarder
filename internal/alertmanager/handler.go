package alertmanager

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "snsalert/internal/aws"
    "snsalert/config"
    "strings"
)

type AlertMessage struct {
    Alerts []struct {
        Status       string            `json:"status"`
        Labels       map[string]string `json:"labels"`
        Annotations  map[string]string `json:"annotations"`
        StartsAt     string            `json:"startsAt"`
        EndsAt       string            `json:"endsAt"`
    } `json:"alerts"`
}

// SNSHandler processes incoming alerts and forwards them to SNS
func SNSHandler(w http.ResponseWriter, r *http.Request) {
    cfg := config.LoadConfig()

    var alertMessage AlertMessage
    if err := json.NewDecoder(r.Body).Decode(&alertMessage); err != nil {
        http.Error(w, "Failed to parse request body", http.StatusBadRequest)
        log.Printf("Error parsing request: %v", err)
        return
    }

    message := formatAlertMessage(alertMessage)

    if err := aws.PublishToSNS(cfg.SNSTopicARN, message); err != nil {
        http.Error(w, fmt.Sprintf("Failed to send message to SNS: %v", err), http.StatusInternalServerError)
        log.Printf("Error sending message to SNS: %v", err)
        return
    }

    log.Printf("Alert sent to SNS topic: %s", cfg.SNSTopicARN)
    fmt.Fprintf(w, "Alert sent to SNS topic: %s", cfg.SNSTopicARN)
}

func formatAlertMessage(alert AlertMessage) string {
    var message strings.Builder
    for _, a := range alert.Alerts {
        fmt.Fprintf(&message, "Alert: %s\n", a.Status)
        fmt.Fprintf(&message, "Start: %s\n", a.StartsAt)
        fmt.Fprintf(&message, "Labels: %v\n", a.Labels)
        fmt.Fprintf(&message, "Annotations: %v\n\n", a.Annotations)
    }
    return message.String()
}
