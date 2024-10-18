package health

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type SNSClient interface {
	CheckSNSConnection(ctx context.Context) error
}

func HealthHandler(w http.ResponseWriter, r *http.Request, snsClient SNSClient) {
	if snsClient == nil {
		log.Error("SNSClient is not initialized")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Проверка соединения с AWS SNS
	err := snsClient.CheckSNSConnection(r.Context())
	if err != nil {
		log.Errorf("Health check failed: %v", err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// Логируем успешное подключение в режиме debug, если ошибок нет
	log.Debug("Successfully connected to AWS SNS during health check")

	// Возвращаем успешный HTTP-ответ
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("healthy")); err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}
