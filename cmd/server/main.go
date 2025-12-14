package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"iot-metrics-service/internal/analytics"
	"iot-metrics-service/internal/buffer"
	"iot-metrics-service/internal/handlers"
	"iot-metrics-service/pkg/redis"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Чтение конфигурации из переменных окружения
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379" // для Docker Compose
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Инициализация зависимостей
	log.Println("🚀 Starting IoT Metrics Service...")
	log.Printf("📡 Redis address: %s", redisAddr)
	log.Printf("🔌 Port: %s", port)

	// 1. Redis клиент
	rdb, ctx := redis.NewClient(redisAddr)

	// 2. Буфер метрик
	metricsBuffer := buffer.NewMetricsBuffer(50)

	// 3. Анализатор
	analyzer := analytics.NewAnalyzer(metricsBuffer)

	// 4. Обработчики
	metricsHandler := handlers.NewMetricsHandler(rdb, metricsBuffer, ctx, analyzer)

	// Создаем базовый обработчик для остальных эндпоинтов
	baseHandler := metricsHandler.BaseHandler

	// Настройка маршрутизатора
	r := mux.NewRouter()

	// API endpoints
	r.HandleFunc("/api/metrics", metricsHandler.MetricsHandler).Methods("POST")
	r.HandleFunc("/api/analyze", baseHandler.AnalyzeHandler).Methods("GET")
	r.HandleFunc("/api/anomalies", baseHandler.AnomaliesHandler).Methods("GET")
	r.HandleFunc("/api/health", baseHandler.HealthHandler).Methods("GET")
	r.Handle("/api/prometheus", promhttp.Handler())

	// Root endpoint
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service": "IoT Metrics Service", "status": "running", "version": "1.0.0"}`))
	}).Methods("GET")

	// Middleware для логирования
	r.Use(loggingMiddleware)

	// Настройка HTTP сервера
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера
	log.Println("✅ Service initialized successfully")
	log.Println("📋 Available endpoints:")
	log.Println("  POST /api/metrics     - Submit IoT metrics")
	log.Println("  GET  /api/analyze     - Get rolling average for device")
	log.Println("  GET  /api/anomalies   - Get detected anomalies")
	log.Println("  GET  /api/health      - Health check")
	log.Println("  GET  /api/prometheus  - Prometheus metrics")
	log.Println("  GET  /                - Service info")

	log.Printf("🌐 Server listening on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

// Middleware для логирования запросов
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем обертку для ResponseWriter для захвата статуса
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.Printf("[%s] %s %s - %d (%v)",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			rw.statusCode,
			duration,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
