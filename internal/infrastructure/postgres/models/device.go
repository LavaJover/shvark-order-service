package models

import "time"

type DeviceModel struct {
	ID 			 string `gorm:"primaryKey"`
	Name 		 string
	TraderID 	 string
	Enabled 	 bool

	DeviceOnline bool

	CreatedAt    time.Time
	UpdatedAt	 time.Time
}