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
