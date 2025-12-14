package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"iot-metrics-service/internal/types"

	"github.com/go-redis/redis/v8"
)

const (
	metricsKey   = "iot:metrics"
	analyticsKey = "iot:analytics"
	windowSize   = 50
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisClient) StoreMetric(metric types.Metric) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	// Store in list (FIFO, max windowSize items)
	pipe := r.client.Pipeline()
	pipe.LPush(r.ctx, metricsKey, data)
	pipe.LTrim(r.ctx, metricsKey, 0, windowSize-1)
	pipe.Expire(r.ctx, metricsKey, 10*time.Minute)

	_, err = pipe.Exec(r.ctx)
	return err
}

func (r *RedisClient) GetRecentMetrics() ([]types.Metric, error) {
	data, err := r.client.LRange(r.ctx, metricsKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	metrics := make([]types.Metric, 0, len(data))
	for _, item := range data {
		var metric types.Metric
		if err := json.Unmarshal([]byte(item), &metric); err == nil {
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

func (r *RedisClient) StoreAnalytics(analytics types.AnalyticsResult) error {
	data, err := json.Marshal(analytics)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, analyticsKey, data, 5*time.Minute).Err()
}

func (r *RedisClient) GetAnalytics() (*types.AnalyticsResult, error) {
	data, err := r.client.Get(r.ctx, analyticsKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var analytics types.AnalyticsResult
	if err := json.Unmarshal([]byte(data), &analytics); err != nil {
		return nil, err
	}

	return &analytics, nil
}

func (r *RedisClient) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
