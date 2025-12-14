package types

// Metric представляет входящую метрику от IoT устройства
type Metric struct {
	Timestamp int64   `json:"timestamp"`
	DeviceID  string  `json:"device_id"`
	CPU       float64 `json:"cpu"`
	RPS       float64 `json:"rps"`
	Memory    float64 `json:"memory"`
}

// AnalyticsResult представляет результат анализа
type AnalyticsResult struct {
	DeviceID       string  `json:"device_id"`
	RollingAverage float64 `json:"rolling_average"`
	ZScore         float64 `json:"z_score"`
	IsAnomaly      bool    `json:"is_anomaly"`
	Timestamp      int64   `json:"timestamp"`
	Value          float64 `json:"value"`
}
