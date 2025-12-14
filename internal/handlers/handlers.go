package handlers

import (
	"context"

	"iot-metrics-service/internal/buffer"
	"iot-metrics-service/internal/types"

	"github.com/go-redis/redis/v8"
)

type BaseHandler struct {
	Redis          *redis.Client
	MetricsBuffer  *buffer.MetricsBuffer
	Ctx            context.Context
	AnomalyChannel chan types.AnalyticsResult
}

func NewBaseHandler(redis *redis.Client, buffer *buffer.MetricsBuffer, ctx context.Context) *BaseHandler {
	return &BaseHandler{
		Redis:          redis,
		MetricsBuffer:  buffer,
		Ctx:            ctx,
		AnomalyChannel: make(chan types.AnalyticsResult, 100),
	}
}

func (h *BaseHandler) GetAnomalyChannel() chan types.AnalyticsResult {
	return h.AnomalyChannel
}
