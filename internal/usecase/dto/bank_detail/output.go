package bankdetaildto

import "github.com/LavaJover/shvark-order-service/internal/domain"

type BankDetailOutput struct {
	domain.BankDetail
}

type Pagination struct {
	CurrentPage int32
	TotalPages	int32
	TotalItems	int32
	ItemsPerPage int32
}

type GetBankDetailsOutput struct {
	BankDetails []*domain.BankDetail
	Pagination
}