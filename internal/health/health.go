package health

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Infof("Error writing response: %v", err)
	}
}
