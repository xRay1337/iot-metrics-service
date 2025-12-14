package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"iot-metrics-service/internal/analytics"
	"iot-metrics-service/internal/buffer"
	"iot-metrics-service/internal/metrics"
	"iot-metrics-service/internal/types"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsHandler struct {
	*BaseHandler
	analyzer *analytics.Analyzer
}

func NewMetricsHandler(redis *redis.Client, buffer *buffer.MetricsBuffer, ctx context.Context, analyzer *analytics.Analyzer) *MetricsHandler {
	baseHandler := NewBaseHandler(redis, buffer, ctx)
	return &MetricsHandler{
		BaseHandler: baseHandler,
		analyzer:    analyzer,
	}
}

func (h *MetricsHandler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues("/metrics"))
	defer timer.ObserveDuration()
	metrics.RequestsTotal.WithLabelValues("/metrics", r.Method).Inc()

	var metric types.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация
	if metric.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}

	// Добавляем метрику в буфер
	h.MetricsBuffer.Add(metric.DeviceID, metric.CPU)

	// Обновляем Prometheus метрики
	metrics.MetricsProcessed.Inc()
	metrics.CurrentRPS.Set(metric.RPS)

	// Кэшируем в Redis
	go h.cacheMetric(metric)

	// Анализируем в отдельной горутине
	go h.analyzeMetric(metric)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Metric received and queued for processing",
	})
}

func (h *MetricsHandler) cacheMetric(metric types.Metric) {
	if h.Redis == nil {
		return
	}

	key := fmt.Sprintf("metric:%s:%d", metric.DeviceID, metric.Timestamp)
	data, _ := json.Marshal(metric)
	h.Redis.Set(h.Ctx, key, data, 10*time.Minute)
}

func (h *MetricsHandler) analyzeMetric(metric types.Metric) {
	result := h.analyzer.Analyze(metric)

	if result.IsAnomaly {
		metrics.AnomaliesDetected.Inc()
		log.Printf("Anomaly detected! Device: %s, CPU: %.2f, Z-Score: %.2f",
			metric.DeviceID, metric.CPU, result.ZScore)
	}

	// Отправляем результат в канал
	select {
	case h.AnomalyChannel <- result:
	default:
		// Канал заполнен, пропускаем
		log.Printf("Anomaly channel is full, skipping result for device %s", metric.DeviceID)
	}
}
