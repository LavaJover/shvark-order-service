package models

import "time"

type DeviceModel struct {
    ID           string `gorm:"primaryKey"`
    Name         string
    TraderID     string `gorm:"index"`
    Enabled      bool
    
    // Статус онлайн
    DeviceOnline bool
    LastPingAt   *time.Time // Время последнего пинга
    LastOnlineAt *time.Time // Время когда был онлайн последний раз
    
    CreatedAt    time.Time
    UpdatedAt    time.Time
}