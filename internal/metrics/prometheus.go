package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iot_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"endpoint", "method"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "iot_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	AnomaliesDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "iot_anomalies_detected_total",
			Help: "Total number of anomalies detected",
		},
	)

	MetricsProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "iot_metrics_processed_total",
			Help: "Total number of IoT metrics processed",
		},
	)

	CurrentRPS = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "iot_current_rps",
			Help: "Current RPS (Requests Per Second) value",
		},
	)
)
