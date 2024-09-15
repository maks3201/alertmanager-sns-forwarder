package ready

import (
	"net/http"

	"github.com/maks3201/sns-alert-service/internal/aws"
	log "github.com/sirupsen/logrus"
)

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	if err := aws.CheckSNSConnection(); err != nil {
		log.Errorf("AWS SNS connection check failed: %v", err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ready")); err != nil {
		log.Errorf("Error writing response: %v", err)
	}
}
