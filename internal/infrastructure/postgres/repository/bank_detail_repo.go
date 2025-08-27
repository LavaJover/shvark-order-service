package repository

import (
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
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
	bankDetailModel := mappers.ToGORMBankDetail(bankDetail)
	return r.DB.Save(bankDetailModel).Error
}

func (r *DefaultBankDetailRepo) CreateBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := mappers.ToGORMBankDetail(bankDetail)
	bankDetailModel.ID = uuid.New().String()

	if err := r.DB.Create(bankDetailModel).Error; err != nil {
		return err
	}

	bankDetail.ID = bankDetailModel.ID

	return nil
}

func (r *DefaultBankDetailRepo) UpdateBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := mappers.ToGORMBankDetail(bankDetail)

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
	return r.DB.Where("id = ?", bankDetailID).Delete(&models.BankDetailModel{}).Error
}

func (r *DefaultBankDetailRepo) GetBankDetailByID(bankDetailID string) (*domain.BankDetail, error) {
	var bankDetailModel models.BankDetailModel
	if err := r.DB.Unscoped().Where("id = ?", bankDetailID).Find(&bankDetailModel).Error; err != nil {
		return nil, err
	}
	return mappers.ToDomainBankDetail(&bankDetailModel), nil
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
		bankDetails[i] = mappers.ToDomainBankDetail(bankDetailModel)
	}

	return bankDetails, total, nil
}

func (r *DefaultBankDetailRepo) FindSuitableBankDetails(searchQuery *domain.SuitablleBankDetailsQuery) ([]*domain.BankDetail, error) {
    var candidates []models.BankDetailModel

    now := time.Now()
    startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

    query := r.DB.
        Model(&models.BankDetailModel{}).
        Where("bank_detail_models.enabled = ?", true).
        Where("min_amount <= ? AND max_amount >= ?", searchQuery.AmountFiat, searchQuery.AmountFiat).
        Where("payment_system = ?", searchQuery.PaymentSystem).
        Where("currency = ?", searchQuery.Currency).
        Where("bank_detail_models.deleted_at IS NULL").
        // Подзапрос для подсчета активных PENDING заявок
        Joins("LEFT JOIN (?) AS pending_orders ON pending_orders.bank_details_id = bank_detail_models.id",
            r.DB.Model(&models.OrderModel{}).
                Select("bank_details_id, COUNT(*) as count").
                Where("status = ?", domain.StatusPending).
                Group("bank_details_id"),
        ).
        // Подзапрос для статистики за день
        Joins("LEFT JOIN (?) AS day_stats ON day_stats.bank_details_id = bank_detail_models.id",
            r.DB.Model(&models.OrderModel{}).
                Select("bank_details_id, COUNT(*) as count, COALESCE(SUM(amount_fiat), 0) as amount").
                Where("status IN (?, ?)", domain.StatusPending, domain.StatusCompleted).
                Where("created_at >= ?", startOfDay).
                Group("bank_details_id"),
        ).
        // Подзапрос для статистики за месяц
        Joins("LEFT JOIN (?) AS month_stats ON month_stats.bank_details_id = bank_detail_models.id",
            r.DB.Model(&models.OrderModel{}).
                Select("bank_details_id, COUNT(*) as count, COALESCE(SUM(amount_fiat), 0) as amount").
                Where("status IN (?, ?)", domain.StatusPending, domain.StatusCompleted).
                Where("created_at >= ?", startOfMonth).
                Group("bank_details_id"),
        ).
        // Подзапрос для времени последней COMPLETED заявки
        Joins("LEFT JOIN (?) AS last_completed ON last_completed.bank_details_id = bank_detail_models.id",
            r.DB.Model(&models.OrderModel{}).
                Select("bank_details_id, MAX(created_at) as last_time").
                Where("status = ?", domain.StatusCompleted).
                Group("bank_details_id"),
        ).
        // Проверка ограничений
        Where("COALESCE(pending_orders.count, 0) < bank_detail_models.max_orders_simultaneosly"). // Максимальное количество активных заявок
        Where("COALESCE(day_stats.count, 0) + 1 <= bank_detail_models.max_quantity_day").         // Дневной лимит количества (+1 для новой заявки)
        Where("COALESCE(day_stats.amount, 0) + ? <= bank_detail_models.max_amount_day", searchQuery.AmountFiat). // Дневной лимит суммы (+новая сумма)
        Where("COALESCE(month_stats.count, 0) + 1 <= bank_detail_models.max_quantity_month").     // Месячный лимит количества (+1 для новой заявки)
        Where("COALESCE(month_stats.amount, 0) + ? <= bank_detail_models.max_amount_month", searchQuery.AmountFiat). // Месячный лимит суммы (+новая сумма)
        Where("last_completed.last_time IS NULL OR last_completed.last_time <= NOW() - (bank_detail_models.delay / 1000000000.0) * INTERVAL '1 SECOND'"). // Задержка после завершенной заявки
        Where("NOT EXISTS (?)", // Проверка на уникальность суммы для активных заявок
            r.DB.Model(&models.OrderModel{}).
                Select("1").
                Where("bank_details_id = bank_detail_models.id").
                Where("status = ?", domain.StatusPending).
                Where("amount_fiat = ?", searchQuery.AmountFiat),
        )

    // Дополнительные фильтры
    if searchQuery.BankCode != "" {
        query = query.Where("bank_detail_models.bank_code = ?", searchQuery.BankCode)
    }

    if searchQuery.NspkCode != "" {
        query = query.Where("bank_detail_models.nspk_code = ?", searchQuery.NspkCode)
    }

    if err := query.Find(&candidates).Error; err != nil {
        return nil, err
    }

    // Преобразование в доменные объекты
    bankDetails := make([]*domain.BankDetail, len(candidates))
    for i, bankDetail := range candidates {
        bankDetails[i] = mappers.ToDomainBankDetail(&bankDetail)
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
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, today).
			Count(&dayCount).Error

		_ = r.DB.Model(&models.OrderModel{}).
			Select("COALESCE(SUM(amount_fiat), 0)").
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, today).
			Scan(&daySum).Error

		// Кол-во и сумма заявок за месяц
		_ = r.DB.Model(&models.OrderModel{}).
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, monthStart).
			Count(&monthCount).Error

		_ = r.DB.Model(&models.OrderModel{}).
			Select("COALESCE(SUM(amount_fiat), 0)").
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, monthStart).
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