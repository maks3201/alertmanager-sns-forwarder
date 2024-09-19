package alertmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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
	cfg           config.Config
	awsClient     aws.SNSClient
	alertChan     chan Alert
	batchMutex    sync.Mutex
	pendingAlerts []Alert
}

func NewHandler(cfg config.Config, awsClient aws.SNSClient) *Handler {
	return &Handler{
		cfg:       cfg,
		awsClient: awsClient,
		alertChan: make(chan Alert, 100),
	}
}

func (h *Handler) SNSHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("Loaded global alertnames: %v", h.cfg.AlertNames)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		log.Errorf("Error reading request body: %v", err)
		return
	}
	defer r.Body.Close()

	var payload AlertmanagerPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
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
			h.alertChan <- alert
		} else {
			log.Infof("Alertname %s is filtered and will not be sent", alertname)
		}
	}

	fmt.Fprintf(w, "Alerts received")
}

func (h *Handler) ProcessBatches(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(h.cfg.BatchWaitSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.sendBatch()
			return
		case alert := <-h.alertChan:
			h.batchMutex.Lock()
			h.pendingAlerts = append(h.pendingAlerts, alert)
			h.batchMutex.Unlock()
		case <-ticker.C:
			h.sendBatch()
		}
	}
}

func (h *Handler) sendBatch() {
	h.batchMutex.Lock()
	if len(h.pendingAlerts) == 0 {
		h.batchMutex.Unlock()
		return
	}

	alertsToSend := h.pendingAlerts
	h.pendingAlerts = nil
	h.batchMutex.Unlock()

	// Group alerts by alertname
	groupedAlerts := groupAlertsByAlertname(alertsToSend)

	// Send each group as a single message
	for alertname, alerts := range groupedAlerts {
		message := formatGroupedAlertMessage(alertname, alerts)

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
			currentTime := time.Now()
			currentTimeParsed, err := parseTime("15:04", currentTime.Format("15:04"))
			if err != nil {
				log.Errorf("Error parsing current time: %v", err)
				continue
			}

			if isTopicAvailable(startTime, endTime, currentTimeParsed) {
				log.Infof("Topic %s is available. Sending batch alert to ARN: %s", topic.Name, topic.ARN)

				if err := h.awsClient.PublishToSNS(context.Background(), topic.ARN, message); err != nil {
					log.Errorf("Error sending batch message to SNS: %v", err)
					continue
				}

				log.Infof("Batch alert sent to SNS topic: %s", topic.ARN)
			} else {
				log.Infof("Topic %s is not available at this time.", topic.Name)
			}
		}
	}
}

func groupAlertsByAlertname(alerts []Alert) map[string][]Alert {
	grouped := make(map[string][]Alert)
	for _, alert := range alerts {
		alertname := alert.Labels["alertname"]
		grouped[alertname] = append(grouped[alertname], alert)
	}
	return grouped
}

// Формирует сообщение для группы алертов, объединяя только поле summary
func formatGroupedAlertMessage(alertname string, alerts []Alert) string {
	var message strings.Builder
	fmt.Fprintf(&message, "[FIRING:%d] SCORUM-GENERAL K8S: %s\n", len(alerts), alertname)

	for _, alert := range alerts {
		// Используем только summary для каждого алерта
		summary := alert.Annotations["summary"]
		if summary != "" {
			fmt.Fprintf(&message, "• %s\n", summary)
		}
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

func isTopicAvailable(startTime, endTime, currentTime time.Time) bool {
	if startTime.Before(endTime) {
		return currentTime.After(startTime) && currentTime.Before(endTime)
	}
	return currentTime.After(startTime) || currentTime.Before(endTime)
}

func parseTime(layout, timeStr string) (time.Time, error) {
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time: %v", err)
	}
	return parsedTime, nil
}
