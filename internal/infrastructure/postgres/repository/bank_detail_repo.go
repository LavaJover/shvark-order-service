package repository

import (
	"fmt"
	"time"

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
		InflowCurrency: bankDetail.InflowCurrency,
		BankCode: bankDetail.BankCode,
		NspkCode: bankDetail.NspkCode,
		DeviceID: bankDetail.DeviceID,
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
		BankCode: bankDetail.BankCode,
		NspkCode: bankDetail.NspkCode,
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
		BankCode: bankDetail.BankCode,
		NspkCode: bankDetail.NspkCode,
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
		BankCode: bankDetailModel.BankCode,
		NspkCode: bankDetailModel.NspkCode,
		CreatedAt: bankDetailModel.CreatedAt,
		UpdatedAt: bankDetailModel.UpdatedAt,
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
			BankCode: bankDetailModel.BankCode,
			NspkCode: bankDetailModel.NspkCode,
			CreatedAt: bankDetailModel.CreatedAt,
			UpdatedAt: bankDetailModel.UpdatedAt,
		}
	}

	return bankDetails, total, nil
}

func (r *DefaultBankDetailRepo) FindSuitableBankDetails(order *domain.Order) ([]*domain.BankDetail, error) {
	var candidates []models.BankDetailModel

	query := r.DB.
		Where("enabled = ?", true).
		Where("min_amount <= ? AND max_amount >= ?", order.AmountFiat, order.AmountFiat).
		Where("payment_system = ?", order.PaymentSystem).
		Where("currency = ?", order.Currency)

	if order.BankCode != "" {
		query = query.Where("bank_code = ?", order.BankCode)
	}

	if order.NspkCode != "" {
		query = query.Where("nspk_code = ?", order.NspkCode)
	}

	if err := query.Find(&candidates).Error; err != nil {
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
			BankCode: bankDetail.BankCode,
			NspkCode: bankDetail.NspkCode,
			CreatedAt: bankDetail.CreatedAt,
			UpdatedAt: bankDetail.UpdatedAt,
		}
	}

	return bankDetails, nil
}

func (r *DefaultBankDetailRepo) GetBankDetailsStatsByTraderID(traderID string) ([]*domain.BankDetailStat, error) {
	var bankDetails []models.BankDetailModel

	if err := r.DB.Where("trader_id = ? AND deleted_at IS NULL", traderID).Find(&bankDetails).Error; err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

	stats := make([]*domain.BankDetailStat, 0, len(bankDetails))

	for _, bd := range bankDetails {
		var dayCount, monthCount int64
		var daySum, monthSum float64

		// Кол-во и сумма заявок за сегодня
		_ = r.DB.Model(&models.OrderModel{}).
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusSucceed), string(domain.StatusCreated)}, today).
			Count(&dayCount).Error

		_ = r.DB.Model(&models.OrderModel{}).
			Select("COALESCE(SUM(amount_fiat), 0)").
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusSucceed), string(domain.StatusCreated)}, today).
			Scan(&daySum).Error

		// Кол-во и сумма заявок за месяц
		_ = r.DB.Model(&models.OrderModel{}).
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusSucceed), string(domain.StatusCreated)}, monthStart).
			Count(&monthCount).Error

		_ = r.DB.Model(&models.OrderModel{}).
			Select("COALESCE(SUM(amount_fiat), 0)").
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusSucceed), string(domain.StatusCreated)}, monthStart).
			Scan(&monthSum).Error

		stats = append(stats, &domain.BankDetailStat{
			BankDetailID:      bd.ID,
			CurrentCountToday: int(dayCount),
			CurrentCountMonth: int(monthCount),
			CurrentAmountToday: daySum,
			CurrentAmountMonth: monthSum,
		})
	}

	return stats, nil
}