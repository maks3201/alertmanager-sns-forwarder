package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"
    "net/http"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sns"
)

// SNSClient is the global variable used for interacting with AWS SNS.
var snsClient *sns.Client

// AlertMessage represents the structure of the alert received from Alertmanager.
type AlertMessage struct {
    Alerts []struct {
        Status      string            `json:"status"`
        Labels      map[string]string `json:"labels"`
        Annotations map[string]string `json:"annotations"`
        StartsAt    string            `json:"startsAt"`
        EndsAt      string            `json:"endsAt"`
    } `json:"alerts"`
}

// initAWSClient initializes the AWS SDK using environment variables
// for connecting to AWS SNS.
func initAWSClient() error {
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
    if err != nil {
        return logJSON("error", fmt.Sprintf("failed to load AWS configuration: %v", err))
    }

    snsClient = sns.NewFromConfig(cfg)

    logJSON("info", fmt.Sprintf("Connected to AWS region: %s", cfg.Region))

    return listSNSTopics()
}

// listSNSTopics lists all available SNS topics.
func listSNSTopics() error {
    _, err := snsClient.ListTopics(context.TODO(), &sns.ListTopicsInput{})
    if err != nil {
        return logJSON("error", fmt.Sprintf("failed to retrieve SNS topics: %v", err))
    }

    return nil
}

// publishToSNS publishes a message to a specific SNS topic.
func publishToSNS(topicArn *string, message *string) error {
    _, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
        TopicArn: topicArn,
        Message:  message,
    })
    if err != nil {
        return fmt.Errorf("failed to send message to SNS: %v", err)
    }
    return nil
}

// snsHandler handles HTTP requests from Alertmanager and sends the alerts to SNS.
func snsHandler(w http.ResponseWriter, r *http.Request) {
    topicArn := os.Getenv("SNS_TOPIC_ARN")
    if topicArn == "" {
        http.Error(w, "SNS_TOPIC_ARN environment variable is missing", http.StatusInternalServerError)
        logJSON("error", "SNS_TOPIC_ARN environment variable is missing")
        return
    }

    var alertMessage AlertMessage
    if err := json.NewDecoder(r.Body).Decode(&alertMessage); err != nil {
        http.Error(w, "failed to parse request body", http.StatusBadRequest)
        logJSON("error", "failed to parse request body")
        return
    }

    message := formatAlertMessage(alertMessage)

    if err := publishToSNS(&topicArn, &message); err != nil {
        http.Error(w, fmt.Sprintf("failed to send message to SNS: %v", err), http.StatusInternalServerError)
        logJSON("error", fmt.Sprintf("failed to send message to SNS: %v", err))
        return
    }

    logJSON("info", fmt.Sprintf("Alert sent to SNS topic: %s", topicArn))
    fmt.Fprintf(w, "Alert sent to SNS topic: %s", topicArn)
}

// formatAlertMessage formats the alert message for sending to SNS.
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

// logJSON creates a structured log message in JSON format with timestamp first.
func logJSON(level string, message string) error {
    logMessage := map[string]string{
        "timestamp": time.Now().Format(time.RFC3339),
        "log.level": level,
        "message":   message,
    }
    logData, err := json.Marshal(logMessage)
    if err != nil {
        return err
    }
    log.Println(string(logData))
    return nil
}

// main is the entry point of the program.
// It initializes the AWS client and starts the HTTP server.
func main() {
    requiredEnvVars := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_ACCOUNT_ID", "AWS_REGION", "SNS_TOPIC_ARN"}
    for _, envVar := range requiredEnvVars {
        if os.Getenv(envVar) == "" {
            logJSON("error", fmt.Sprintf("%s environment variable is missing", envVar))
            log.Fatal(fmt.Sprintf("%s environment variable is missing", envVar))
        }
    }

    if err := initAWSClient(); err != nil {
        log.Fatalf("failed to initialize AWS client: %v", err)
    }

    http.HandleFunc("/alert", snsHandler)

    logJSON("info", "Server started on port 80")
    if err := http.ListenAndServe(":80", nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
