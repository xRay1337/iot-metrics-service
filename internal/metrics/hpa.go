package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Метрики для Kubernetes HPA (Horizontal Pod Autoscaler)
	PodCPUUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "iot_pod_cpu_usage_percent",
			Help: "Current CPU usage percentage of the pod",
		},
	)

	PodMemoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "iot_pod_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)

	GoRoutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "iot_goroutines_count",
			Help: "Number of active goroutines",
		},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "iot_active_connections",
			Help: "Number of active HTTP connections",
		},
	)
)

// StartHPAMetrics начинает сбор метрик для HPA
func StartHPAMetrics() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			UpdateHPAMetrics()
		}
	}()
}

// UpdateHPAMetrics обновляет метрики для HPA
func UpdateHPAMetrics() {
	// 1. Обновляем количество горутин
	GoRoutines.Set(float64(runtime.NumGoroutine()))

	// 2. Обновляем использование памяти
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	PodMemoryUsage.Set(float64(m.Alloc))

	// 3. CPU usage можно симулировать или использовать gopsutil
	// Для простоты - симулируем на основе нагрузки
	simulateCPUUsage()
}

func simulateCPUUsage() {
	// Простая симуляция CPU usage на основе количества горутин
	goroutines := runtime.NumGoroutine()
	cpuUsage := float64(goroutines) * 0.5 // Каждая горутина ~0.5% CPU
	if cpuUsage > 100 {
		cpuUsage = 100
	}
	PodCPUUsage.Set(cpuUsage)
}
