package domain

type BankDetailRepository interface {
	SaveBankDetail(bankDetail *BankDetail) error 
}