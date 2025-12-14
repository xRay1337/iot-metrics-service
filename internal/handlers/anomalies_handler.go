package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"iot-metrics-service/internal/metrics"
	"iot-metrics-service/internal/types"
)

// AnomaliesHandler возвращает список обнаруженных аномалий
func (h *BaseHandler) AnomaliesHandler(w http.ResponseWriter, r *http.Request) {
	metrics.RequestsTotal.WithLabelValues("/anomalies", r.Method).Inc()

	anomalies := make([]types.AnalyticsResult, 0)
	timeout := time.After(100 * time.Millisecond)

	// Собираем аномалии из канала
drainLoop:
	for {
		select {
		case anomaly := <-h.AnomalyChannel:
			anomalies = append(anomalies, anomaly)
		case <-timeout:
			break drainLoop
		default:
			break drainLoop
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(anomalies),
		"anomalies": anomalies,
	})
}
