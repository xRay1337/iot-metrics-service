package handlers

import (
	"encoding/json"
	"net/http"

	"iot-metrics-service/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// AnalyzeHandler возвращает результаты анализа для устройства
func (h *BaseHandler) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("/analyze"))
	defer timer.ObserveDuration()
	metrics.RequestsTotal.WithLabelValues("/analyze", r.Method).Inc()

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, "device_id parameter is required", http.StatusBadRequest)
		return
	}

	rollingAvg := h.MetricsBuffer.GetRollingAverage(deviceID)

	response := map[string]interface{}{
		"device_id":       deviceID,
		"rolling_average": rollingAvg,
		"window_size":     50,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
