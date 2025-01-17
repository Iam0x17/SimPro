package handler

import (
	"SimPro/services"
	"encoding/json"
	"fmt"
	"net/http"
)

type ServiceControlRequest struct {
	ServiceName string `json:"service_name"`
	Action      string `json:"action"`
}

func ServiceControlHandler(w http.ResponseWriter, r *http.Request) {
	var req ServiceControlRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	serviceName := req.ServiceName
	action := req.Action

	manager := services.GetServiceManager()

	switch action {
	case "start":
		err = manager.StartServiceByName(serviceName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to start service %s: %v", serviceName, err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Service %s started", serviceName)))
	case "stop":
		err = manager.StopServiceByName(serviceName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to stop service %s: %v", serviceName, err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Service %s stopped", serviceName)))
	case "status":
		var status string
		status, err = manager.GetServiceStatusByName(serviceName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to stop service %s: %v", serviceName, err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Service %s status is %s", serviceName, status)))
	default:
		http.Error(w, "Invalid action. Use 'start' or 'stop'", http.StatusBadRequest)
	}
}
