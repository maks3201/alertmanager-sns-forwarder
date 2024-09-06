package alertmanager

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
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

    currentTime := time.Now().Format("15:04")

    for _, topic := range cfg.Topics {
        startTime := config.ParseTime("15:04", topic.StartTime)
        endTime := config.ParseTime("15:04", topic.EndTime)
        currentTimeParsed := config.ParseTime("15:04", currentTime)

        if isTopicAvailable(startTime, endTime, currentTimeParsed) {
            log.Printf("Topic %s is available. Sending alert...", topic.Name)
            var payload AlertmanagerPayload
            if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
                http.Error(w, "Failed to parse request body", http.StatusBadRequest)
                log.Printf("Error parsing request: %v", err)
                return
            }

            message := formatAlertMessage(payload)

            log.Printf("Sending message to SNS topic ARN: %s", topic.ARN)

            if err := aws.PublishToSNS(topic.ARN, message); err != nil {
                http.Error(w, fmt.Sprintf("Failed to send message to SNS: %v", err), http.StatusInternalServerError)
                log.Printf("Error sending message to SNS: %v", err)
                return
            }

            log.Printf("Alert sent to SNS topic: %s", topic.ARN)
            fmt.Fprintf(w, "Alert sent to SNS topic: %s", topic.ARN)
        } else {
            log.Printf("Topic %s is not available at this time.", topic.Name)
        }
    }
}

func isTopicAvailable(startTime, endTime, currentTime time.Time) bool {
    if startTime.Before(endTime) {
        return currentTime.After(startTime) && currentTime.Before(endTime)
    }
    return currentTime.After(startTime) || currentTime.Before(endTime)
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
