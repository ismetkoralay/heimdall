// Package health provides health check functionality for the Heimdall service. It defines the HealthCheck interface and implements a basic health check mechanism to monitor the status of the service.
package health

import (
	"encoding/json"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
