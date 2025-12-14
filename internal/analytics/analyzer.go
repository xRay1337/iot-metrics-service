package analytics

import (
	"math"

	"iot-metrics-service/internal/buffer"
	"iot-metrics-service/internal/types"
)

type Analyzer struct {
	buffer *buffer.MetricsBuffer
}

func NewAnalyzer(buffer *buffer.MetricsBuffer) *Analyzer {
	return &Analyzer{
		buffer: buffer,
	}
}

func (a *Analyzer) Analyze(metric types.Metric) types.AnalyticsResult {
	rollingAvg := a.buffer.GetRollingAverage(metric.DeviceID)
	zScore := a.buffer.GetZScore(metric.DeviceID, metric.CPU)

	// Порог для аномалий: |z-score| > 2
	isAnomaly := math.Abs(zScore) > 2.0

	return types.AnalyticsResult{
		DeviceID:       metric.DeviceID,
		RollingAverage: rollingAvg,
		ZScore:         zScore,
		IsAnomaly:      isAnomaly,
		Timestamp:      metric.Timestamp,
		Value:          metric.CPU,
	}
}
