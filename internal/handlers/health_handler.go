package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler проверка здоровья сервиса
func (h *BaseHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	}

	// Проверяем Redis
	if h.Redis != nil {
		_, err := h.Redis.Ping(h.Ctx).Result()
		if err != nil {
			health["redis"] = "disconnected"
			health["status"] = "degraded"
		} else {
			health["redis"] = "connected"
		}
	} else {
		health["redis"] = "not_initialized"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
