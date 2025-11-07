package domain

import "time"

type AutomaticLogFilter struct {
    DeviceID  string
    TraderID  string
    Success   *bool
    Action    string
    StartDate time.Time
    EndDate   time.Time
    Limit     int
    Offset    int
}

type AutomaticLog struct {
    ID             string
    DeviceID       string
    TraderID       string
    OrderID        string
    Amount         float64
    PaymentSystem  string
    Direction      string
    Methods        []string
    ReceivedAt     time.Time
    Text           string
    Action         string
    Success        bool
    OrdersFound    int
    ErrorMessage   string
    ProcessingTime int64
    BankName       string
    CardNumber     string
    CreatedAt      time.Time
}

type AutomaticStats struct {
    TotalAttempts      int64                  `json:"total_attempts"`
    SuccessfulAttempts int64                  `json:"successful_attempts"`
    ApprovedOrders     int64                  `json:"approved_orders"`
    NotFoundCount      int64                  `json:"not_found_count"`
    FailedCount        int64                  `json:"failed_count"`
    AvgProcessingTime  float64                `json:"avg_processing_time"`
    DeviceStats        map[string]DeviceStats `json:"device_stats"`
}

type DeviceStats struct {
    TotalAttempts int64   `json:"total_attempts"`
    SuccessCount  int64   `json:"success_count"`
    SuccessRate   float64 `json:"success_rate"`
}

// CalculateSuccessRate вычисляет процент успешных операций
func (ds *DeviceStats) CalculateSuccessRate() {
    if ds.TotalAttempts > 0 {
        ds.SuccessRate = float64(ds.SuccessCount) / float64(ds.TotalAttempts) * 100
    }
}