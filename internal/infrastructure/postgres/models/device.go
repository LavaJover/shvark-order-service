package models

type DeviceModel struct {
	ID 			string `gorm:"primaryKey"`
	Name 		string
	TraderID 	string
	Enabled 	bool
}