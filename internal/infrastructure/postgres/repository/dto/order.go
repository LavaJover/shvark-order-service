package dto

// ExpiredOrderData - структура для возврата данных об отмененных заказах
type ExpiredOrderData struct {
    ID                  string  `db:"id"`
    MerchantID          string  `db:"merchant_id"`
    AmountFiat          float64 `db:"amount_fiat"`
    Currency            string  `db:"currency"`
    CallbackURL         string  `db:"callback_url"`
    MerchantOrderID     string  `db:"merchant_order_id"`
    TraderRewardPercent float64 `db:"trader_reward_percent"`
    PlatformFee         float64 `db:"platform_fee"`
    TraderID            string  `db:"trader_id"`
    BankName            string  `db:"bank_name"`
    Phone               string  `db:"phone"`
    CardNumber          string  `db:"card_number"`
    Owner               string  `db:"owner"`
}