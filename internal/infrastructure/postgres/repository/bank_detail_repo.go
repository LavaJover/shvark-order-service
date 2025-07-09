package repository

import (
	"fmt"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DefaultBankDetailRepo struct {
	DB *gorm.DB
}

func NewDefaultBankDetailRepo(db *gorm.DB) *DefaultBankDetailRepo {
	return &DefaultBankDetailRepo{DB: db}
}

func (r *DefaultBankDetailRepo) SaveBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := models.BankDetailModel{
		ID: bankDetail.ID,
		TraderID: bankDetail.TraderID,
		Country: bankDetail.Country,
		Currency: bankDetail.Currency,
		MinAmount: bankDetail.MinAmount,
		MaxAmount: bankDetail.MaxAmount,
		BankName: bankDetail.BankName,
		PaymentSystem: bankDetail.PaymentSystem,
		Delay: bankDetail.Delay,
		Enabled: bankDetail.Enabled,
		CardNumber: bankDetail.CardNumber,
		Phone: bankDetail.Phone,
		Owner: bankDetail.Owner,
		MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
		MaxAmountDay: bankDetail.MaxAmountDay,
		MaxAmountMonth: bankDetail.MaxAmountMonth,
		MaxQuantityDay: bankDetail.MaxQuantityDay,
		MaxQuantityMonth: bankDetail.MaxQuantityMonth,
	}

	return r.DB.Save(&bankDetailModel).Error
}

func (r *DefaultBankDetailRepo) CreateBankDetail(bankDetail *domain.BankDetail) (string, error) {
	bankDetailModel := &models.BankDetailModel{
		ID: uuid.New().String(),
		TraderID: bankDetail.TraderID,
		Country: bankDetail.Country,
		Currency: bankDetail.Currency,
		InflowCurrency: bankDetail.InflowCurrency,
		MinAmount: bankDetail.MinAmount,
		MaxAmount: bankDetail.MaxAmount,
		BankName: bankDetail.BankName,
		PaymentSystem: bankDetail.PaymentSystem,
		Delay: bankDetail.Delay,
		Enabled: bankDetail.Enabled,
		CardNumber: bankDetail.CardNumber,
		Phone: bankDetail.Phone,
		Owner: bankDetail.Owner,
		MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
		MaxAmountDay: bankDetail.MaxAmountDay,
		MaxAmountMonth: bankDetail.MaxAmountMonth,
		MaxQuantityDay: bankDetail.MaxQuantityDay,
		MaxQuantityMonth: bankDetail.MaxAmountMonth,
		DeviceID: bankDetail.DeviceID,
	}

	if err := r.DB.Create(bankDetailModel).Error; err != nil {
		return "", err
	}

	bankDetail.ID = bankDetailModel.ID

	return bankDetail.ID, nil
}

func (r *DefaultBankDetailRepo) UpdateBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := &models.BankDetailModel{
		ID: bankDetail.ID,
		TraderID: bankDetail.TraderID,
		Country: bankDetail.Country,
		Currency: bankDetail.Currency,
		MinAmount: bankDetail.MinAmount,
		MaxAmount: bankDetail.MaxAmount,
		BankName: bankDetail.BankName,
		PaymentSystem: bankDetail.PaymentSystem,
		Delay: bankDetail.Delay,
		Enabled: bankDetail.Enabled,
		CardNumber: bankDetail.CardNumber,
		Phone: bankDetail.Phone,
		Owner: bankDetail.Owner,
		MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
		MaxAmountDay: bankDetail.MaxAmountDay,
		MaxAmountMonth: bankDetail.MaxAmountMonth,
		MaxQuantityDay: bankDetail.MaxQuantityDay,
		MaxQuantityMonth: bankDetail.MaxAmountMonth,
		InflowCurrency: bankDetail.InflowCurrency,
		DeviceID: bankDetail.DeviceID,
	}

	if err := r.DB.Model(&models.BankDetailModel{}).Where("id = ?", bankDetailModel.ID).Updates(bankDetailModel).Error; err != nil {
		return err
	}

	if err := r.DB.Model(&models.BankDetailModel{}).Where("id = ?", bankDetailModel.ID).Updates(map[string]interface{}{
		"enabled": bankDetail.Enabled,
		"delay": bankDetail.Delay,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultBankDetailRepo) DeleteBankDetail(bankDetailID string) error {
	return r.DB.Delete(&models.BankDetailModel{ID: bankDetailID}).Error
}

func (r *DefaultBankDetailRepo) GetBankDetailByID(bankDetailID string) (*domain.BankDetail, error) {
	var bankDetailModel models.BankDetailModel
	if err := r.DB.Where("id = ?", bankDetailID).Find(&bankDetailModel).Error; err != nil {
		return nil, err
	}
	return &domain.BankDetail{
		ID: bankDetailModel.ID,
		TraderID: bankDetailModel.TraderID,
		Country: bankDetailModel.Country,
		Currency: bankDetailModel.Currency,
		InflowCurrency: bankDetailModel.InflowCurrency,
		MinAmount: bankDetailModel.MinAmount,
		MaxAmount: bankDetailModel.MaxAmount,
		BankName: bankDetailModel.BankName,
		PaymentSystem: bankDetailModel.PaymentSystem,
		Delay: bankDetailModel.Delay,
		Enabled: bankDetailModel.Enabled,
		CardNumber: bankDetailModel.CardNumber,
		Phone: bankDetailModel.Phone,
		Owner: bankDetailModel.Owner,
		MaxOrdersSimultaneosly: bankDetailModel.MaxOrdersSimultaneosly,
		MaxAmountDay: bankDetailModel.MaxAmountDay,
		MaxAmountMonth: bankDetailModel.MaxAmountMonth,
		MaxQuantityDay: bankDetailModel.MaxQuantityDay,
		MaxQuantityMonth: bankDetailModel.MaxQuantityMonth,
		DeviceID: bankDetailModel.DeviceID,
	}, nil
}

func (r *DefaultBankDetailRepo) GetBankDetailsByTraderID(
	traderID string, 
	page, limit int,
	sortOrder, sortBy string,
) ([]*domain.BankDetail, int64, error) {
	var bankDetailModels []*models.BankDetailModel
	var total int64
	
	safeSortBy := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"min_amount": true,
		"max_amount": true,
	}

	if !safeSortBy[sortBy] {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 1
	}

	offset := (page - 1) * limit
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	err := r.DB.
		Model(&models.BankDetailModel{}).
		Where("trader_id = ?", traderID).
		Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&bankDetailModels).Error
	
	if err != nil {
		return nil, 0, err
	}

	if err := r.DB.Model(&models.BankDetailModel{}).Where("trader_id = ?", traderID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	bankDetails := make([]*domain.BankDetail, len(bankDetailModels))
	for i, bankDetailModel := range bankDetailModels{
		bankDetails[i] = &domain.BankDetail{
			ID: bankDetailModel.ID,
			TraderID: bankDetailModel.TraderID,
			Country: bankDetailModel.Country,
			Currency: bankDetailModel.Currency,
			InflowCurrency: bankDetailModel.InflowCurrency,
			MinAmount: bankDetailModel.MinAmount,
			MaxAmount: bankDetailModel.MaxAmount,
			BankName: bankDetailModel.BankName,
			PaymentSystem: bankDetailModel.PaymentSystem,
			Delay: bankDetailModel.Delay,
			Enabled: bankDetailModel.Enabled,
			CardNumber: bankDetailModel.CardNumber,
			Phone: bankDetailModel.Phone,
			Owner: bankDetailModel.Owner,
			MaxOrdersSimultaneosly: bankDetailModel.MaxOrdersSimultaneosly,
			MaxAmountDay: bankDetailModel.MaxAmountDay,
			MaxAmountMonth: bankDetailModel.MaxAmountMonth,
			MaxQuantityDay: bankDetailModel.MaxQuantityDay,
			MaxQuantityMonth: bankDetailModel.MaxQuantityMonth,
			DeviceID: bankDetailModel.DeviceID,
			CreatedAt: bankDetailModel.CreatedAt,
			UpdatedAt: bankDetailModel.UpdatedAt,
		}
	}

	return bankDetails, total, nil
}

func (r *DefaultBankDetailRepo) FindSuitableBankDetails(order *domain.Order) ([]*domain.BankDetail, error) {
	var candidates []models.BankDetailModel

	if err := r.DB.
		Where("enabled = ?", true).
		Where("min_amount <= ? AND max_amount >= ?", order.AmountFiat, order.AmountFiat).
		Where("payment_system = ?", order.PaymentSystem).
		Where("currency = ?", order.Currency).
		Find(&candidates).Error; err != nil {
		return nil, err
	}

	bankDetails := make([]*domain.BankDetail, len(candidates))
	for i, bankDetail := range candidates {
		bankDetails[i] = &domain.BankDetail{
			ID: bankDetail.ID,
			TraderID: bankDetail.TraderID,
			Country: bankDetail.Country,
			Currency: bankDetail.Currency,
			InflowCurrency: bankDetail.InflowCurrency,
			MinAmount: bankDetail.MinAmount,
			MaxAmount: bankDetail.MaxAmount,
			BankName: bankDetail.BankName,
			PaymentSystem: bankDetail.PaymentSystem,
			Delay: bankDetail.Delay,
			Enabled: bankDetail.Enabled,
			CardNumber: bankDetail.CardNumber,
			Phone: bankDetail.Phone,
			Owner: bankDetail.Owner,
			MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
			MaxAmountDay: bankDetail.MaxAmountDay,
			MaxAmountMonth: bankDetail.MaxAmountMonth,
			MaxQuantityDay: bankDetail.MaxQuantityDay,
			MaxQuantityMonth: bankDetail.MaxQuantityMonth,
			DeviceID: bankDetail.DeviceID,
			CreatedAt: bankDetail.CreatedAt,
			UpdatedAt: bankDetail.UpdatedAt,
		}
	}

	return bankDetails, nil
}