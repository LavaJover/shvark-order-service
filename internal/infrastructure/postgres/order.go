package postgres

type OrderModel struct {
	ID 			  string  			`gorm:"primaryKey;type:uuid"`
	MerchantID 	  string  			`gorm:"type:uuid"`
	Amount 		  float32	
	Currency 	  string		
	Country 	  string
	ClientEmail   string
	MetadataJSON  string
	Status 		  string
	PaymentSystem string
	BankDetailsID string  			`gorm:"type:uuid"`	
	BankDetail 	  BankDetailModel   `gorm:"foreignKey:BankDetailsID;references:ID;constraint:OnUpdate,OnDelete:Restrict;"`
}