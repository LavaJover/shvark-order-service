package merchantdto

import "time"

type CreateStoreInput struct {
    MerchantID    string        `json:"merchant_id" validate:"required"`
    Name          string        `json:"name" validate:"required"`
    PlatformFee   float64       `json:"platform_fee" validate:"min=0,max=100"`
    DealsDuration time.Duration `json:"deals_duration"`
    Description   string        `json:"description"`
    Category      string        `json:"category"`
    MaxDailyDeals int           `json:"max_daily_deals"`
    MinDealAmount float64       `json:"min_deal_amount"`
    MaxDealAmount float64       `json:"max_deal_amount"`
    Currency      string        `json:"currency"`
}

type UpdateStoreInput struct {
    ID            string        `json:"id" validate:"required"`
    Name          string        `json:"name"`
    PlatformFee   float64       `json:"platform_fee" validate:"min=0,max=100"`
    DealsDuration time.Duration `json:"deals_duration"`
    Description   string        `json:"description"`
    Category      string        `json:"category"`
    MaxDailyDeals int           `json:"max_daily_deals"`
    MinDealAmount float64       `json:"min_deal_amount"`
    MaxDealAmount float64       `json:"max_deal_amount"`
    Currency      string        `json:"currency"`
    IsActive      bool          `json:"is_active"`
}