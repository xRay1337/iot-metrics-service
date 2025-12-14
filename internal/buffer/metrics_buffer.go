package buffer

import (
	"math"
	"sync"
)

// MetricsBuffer хранит метрики для анализа
type MetricsBuffer struct {
	mu      sync.RWMutex
	data    map[string][]float64
	window  int
	maxSize int
}

func NewMetricsBuffer(window int) *MetricsBuffer {
	return &MetricsBuffer{
		data:    make(map[string][]float64),
		window:  window,
		maxSize: 1000,
	}
}

func (mb *MetricsBuffer) Add(deviceID string, value float64) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if _, exists := mb.data[deviceID]; !exists {
		mb.data[deviceID] = make([]float64, 0, mb.maxSize)
	}

	mb.data[deviceID] = append(mb.data[deviceID], value)

	// Ограничиваем размер буфера
	if len(mb.data[deviceID]) > mb.maxSize {
		mb.data[deviceID] = mb.data[deviceID][len(mb.data[deviceID])-mb.maxSize:]
	}
}

func (mb *MetricsBuffer) GetRollingAverage(deviceID string) float64 {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	values, exists := mb.data[deviceID]
	if !exists || len(values) == 0 {
		return 0
	}

	// Вычисляем скользящее среднее по последним N значениям
	start := 0
	if len(values) > mb.window {
		start = len(values) - mb.window
	}

	sum := 0.0
	count := 0
	for i := start; i < len(values); i++ {
		sum += values[i]
		count++
	}

	if count == 0 {
		return 0
	}

	return sum / float64(count)
}

func (mb *MetricsBuffer) GetZScore(deviceID string, currentValue float64) float64 {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	values, exists := mb.data[deviceID]
	if !exists || len(values) < 2 {
		return 0
	}

	// Вычисляем среднее и стандартное отклонение
	start := 0
	if len(values) > mb.window {
		start = len(values) - mb.window
	}

	var sum float64
	count := 0
	for i := start; i < len(values); i++ {
		sum += values[i]
		count++
	}

	if count == 0 {
		return 0
	}

	mean := sum / float64(count)

	// Стандартное отклонение
	var variance float64
	for i := start; i < len(values); i++ {
		diff := values[i] - mean
		variance += diff * diff
	}
	variance /= float64(count)
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	// Z-score
	zScore := (currentValue - mean) / stdDev
	return zScore
}
