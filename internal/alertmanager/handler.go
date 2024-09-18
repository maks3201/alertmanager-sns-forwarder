package alertmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/maks3201/sns-alert-service/config"
	"github.com/maks3201/sns-alert-service/internal/aws"
	log "github.com/sirupsen/logrus"
)

type AlertmanagerPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

type Handler struct {
	cfg       config.Config
	awsClient aws.SNSClient
}

func NewHandler(cfg config.Config, awsClient aws.SNSClient) *Handler {
	return &Handler{
		cfg:       cfg,
		awsClient: awsClient,
	}
}

func (h *Handler) SNSHandler(w http.ResponseWriter, r *http.Request) {

	log.Infof("Loaded global alertnames: %v", h.cfg.AlertNames)

	currentTime := time.Now().Format("15:04")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		log.Errorf("Error reading request body: %v", err)
		return
	}
	defer r.Body.Close()

	for _, topic := range h.cfg.Topics {
		startTime, err := parseTime("15:04", topic.StartTime)
		if err != nil {
			log.Errorf("Error parsing start time: %v", err)
			continue
		}
		endTime, err := parseTime("15:04", topic.EndTime)
		if err != nil {
			log.Errorf("Error parsing end time: %v", err)
			continue
		}
		currentTimeParsed, err := parseTime("15:04", currentTime)
		if err != nil {
			log.Errorf("Error parsing current time: %v", err)
			continue
		}

		if isTopicAvailable(startTime, endTime, currentTimeParsed) {
			log.Infof("Topic %s is available. Sending alert to ARN: %s", topic.Name, topic.ARN)

			bodyReader := io.NopCloser(bytes.NewBuffer(bodyBytes))

			var payload AlertmanagerPayload
			if err := json.NewDecoder(bodyReader).Decode(&payload); err != nil {
				http.Error(w, "Failed to parse request body", http.StatusBadRequest)
				log.Errorf("Error parsing request: %v", err)
				return
			}

			for _, alert := range payload.Alerts {
				alertname := alert.Labels["alertname"]
				log.Infof("Received alertname: %s", alertname)
				log.Infof("Allowed alertnames: %v", h.cfg.AlertNames)

				if isAlertFiltered(alertname, h.cfg.AlertNames) {
					log.Infof("Alertname %s is allowed", alertname)
					message := formatAlertMessage(payload)

					if err := h.awsClient.PublishToSNS(r.Context(), topic.ARN, message); err != nil {
						log.Errorf("Error sending message to SNS: %v", err)
						http.Error(w, fmt.Sprintf("Failed to send message to SNS: %v", err), http.StatusInternalServerError)
						continue
					}

					log.Infof("Alert sent to SNS topic: %s", topic.ARN)
				} else {
					log.Infof("Alertname %s is filtered and will not be sent", alertname)
				}
			}
		} else {
			log.Infof("Topic %s is not available at this time.", topic.Name)
		}
	}

	fmt.Fprintf(w, "Alerts sent to all available topics")
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

func isAlertFiltered(alertname string, allowedAlertNames []string) bool {
	if len(allowedAlertNames) == 0 {
		return true
	}
	for _, allowedAlert := range allowedAlertNames {
		if allowedAlert == alertname {
			return true
		}
	}
	return false
}

func parseTime(layout, timeStr string) (time.Time, error) {
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time: %v", err)
	}
	return parsedTime, nil
}
