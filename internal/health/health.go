package health

import (
    "net/http"
    "log"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    if _, err := w.Write([]byte("ok")); err != nil {
        log.Printf("Error writing response: %v", err)
    }
}
