package merchantdto

import "time"

type StoreResponse struct {
    ID            string        `json:"id"`
    MerchantID    string        `json:"merchant_id"`
    Name          string        `json:"name"`
    PlatformFee   float64       `json:"platform_fee"`
    IsActive      bool          `json:"is_active"`
    DealsDuration time.Duration `json:"deals_duration"`
    Description   string        `json:"description"`
    Category      string        `json:"category"`
    MaxDailyDeals int           `json:"max_daily_deals"`
    MinDealAmount float64       `json:"min_deal_amount"`
    MaxDealAmount float64       `json:"max_deal_amount"`
    Currency      string        `json:"currency"`
    CreatedAt     time.Time     `json:"created_at"`
    UpdatedAt     time.Time     `json:"updated_at"`
}

type StoreListResponse struct {
    Stores []StoreResponse `json:"stores"`
    Total  int             `json:"total"`
    Page   int32           `json:"page"`
    Limit  int32           `json:"limit"`
}