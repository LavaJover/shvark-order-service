package domain

import "time"

type BankDetail struct {
	ID 				string
	TraderID 		string
	Country 		string
	Currency 		string
	MinAmount 		float32
	MaxAmount 		float32
	BankName 		string
	PaymentSystem 	string
	Delay			time.Duration
	Enabled 		bool
}

type BankDetailQuery struct {
	Amount 			float32
	Currency 		string
	PaymentSystem 	string
	Country 		string
}