package repository

import (
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
    // Этап 1: Быстрая предварительная фильтрация по статическим параметрам
    baseCandidates, err := r.findBaseCandidates(searchQuery)
    if err != nil {
        return nil, err
    }
    
    if len(baseCandidates) == 0 {
        return []*domain.BankDetail{}, nil
    }

    // Этап 2: Проверка динамических ограничений с использованием денормализации
    finalCandidates, err := r.applyDynamicConstraintsOptimized(baseCandidates, searchQuery)
    if err != nil {
        return nil, err
    }

    return finalCandidates, nil
}

// Этап 1 остается без изменений
func (r *DefaultBankDetailRepo) findBaseCandidates(searchQuery *domain.SuitablleBankDetailsQuery) ([]models.BankDetailModel, error) {
    var baseCandidates []models.BankDetailModel
    
    query := r.DB.Model(&models.BankDetailModel{}).
        Where("enabled = ?", true).
        Where("min_amount <= ? AND max_amount >= ?", searchQuery.AmountFiat, searchQuery.AmountFiat).
        Where("payment_system = ?", searchQuery.PaymentSystem).
        Where("currency = ?", searchQuery.Currency).
        Where("deleted_at IS NULL")
    
    if searchQuery.BankCode != "" {
        query = query.Where("bank_code = ?", searchQuery.BankCode)
    }
    
    if searchQuery.NspkCode != "" {
        query = query.Where("nspk_code = ?", searchQuery.NspkCode)
    }
    
    err := query.Select("id, trader_id, country, currency, inflow_currency, " +
        "min_amount, max_amount, bank_name, bank_code, nspk_code, " +
        "payment_system, delay, enabled, card_number, phone, owner, " +
        "max_orders_simultaneosly, max_amount_day, max_amount_month, " +
        "max_quantity_day, max_quantity_month, device_id, created_at, updated_at").
        Find(&baseCandidates).Error
        
    return baseCandidates, err
}

// КАРДИНАЛЬНО ОПТИМИЗИРОВАННЫЙ этап 2 с использованием денормализации
func (r *DefaultBankDetailRepo) applyDynamicConstraintsOptimized(baseCandidates []models.BankDetailModel, searchQuery *domain.SuitablleBankDetailsQuery) ([]*domain.BankDetail, error) {
    if len(baseCandidates) == 0 {
        return []*domain.BankDetail{}, nil
    }
    
    // Получаем trader_id кандидатов для прямого поиска по order_models
    traderIDs := make([]string, len(baseCandidates))
    traderToBankDetail := make(map[string]models.BankDetailModel)
    
    for i, candidate := range baseCandidates {
        traderIDs[i] = candidate.TraderID
        traderToBankDetail[candidate.TraderID] = candidate
    }
    
    now := time.Now()
    startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
    
    // СУПЕР ОПТИМИЗИРОВАННЫЙ запрос БЕЗ JOIN'ов - только к order_models!
    sqlQuery := `
        WITH trader_stats AS (
            SELECT 
                trader_id,
                -- Количество активных заказов
                SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as pending_count,
                -- Статистика за день
                SUM(CASE WHEN status IN (?, ?) AND created_at >= ? THEN 1 ELSE 0 END) as day_count,
                SUM(CASE WHEN status IN (?, ?) AND created_at >= ? THEN amount_fiat ELSE 0 END) as day_amount,
                -- Статистика за месяц
                SUM(CASE WHEN status IN (?, ?) AND created_at >= ? THEN 1 ELSE 0 END) as month_count,
                SUM(CASE WHEN status IN (?, ?) AND created_at >= ? THEN amount_fiat ELSE 0 END) as month_amount,
                -- Время последнего завершенного заказа
                MAX(CASE WHEN status = ? THEN created_at END) as last_completed_time,
                -- Проверка на дублирующие заказы
                SUM(CASE WHEN status = ? AND amount_fiat = ? THEN 1 ELSE 0 END) as duplicate_count
            FROM order_models 
            WHERE trader_id = ANY(?)
            GROUP BY trader_id
        )
        SELECT DISTINCT bd.* 
        FROM bank_detail_models bd
        LEFT JOIN trader_stats ts ON bd.trader_id = ts.trader_id
        WHERE bd.trader_id = ANY(?)
          AND bd.enabled = true
          AND bd.deleted_at IS NULL
          AND COALESCE(ts.pending_count, 0) < bd.max_orders_simultaneosly
          AND COALESCE(ts.day_count, 0) + 1 <= bd.max_quantity_day
          AND COALESCE(ts.day_amount, 0) + ? <= bd.max_amount_day
          AND COALESCE(ts.month_count, 0) + 1 <= bd.max_quantity_month
          AND COALESCE(ts.month_amount, 0) + ? <= bd.max_amount_month
          AND (ts.last_completed_time IS NULL OR ts.last_completed_time <= NOW() - (bd.delay / 1000000000.0) * INTERVAL '1 SECOND')
          AND COALESCE(ts.duplicate_count, 0) = 0
    `
    
    var finalCandidates []models.BankDetailModel
    
    // Выполняем СУПЕР оптимизированный запрос
    err := r.DB.Raw(sqlQuery,
        domain.StatusPending, // pending_count
        domain.StatusPending, domain.StatusCompleted, startOfDay, // day_count
        domain.StatusPending, domain.StatusCompleted, startOfDay, // day_amount  
        domain.StatusPending, domain.StatusCompleted, startOfMonth, // month_count
        domain.StatusPending, domain.StatusCompleted, startOfMonth, // month_amount
        domain.StatusCompleted, // last_completed_time
        domain.StatusPending, searchQuery.AmountFiat, // duplicate_count
        pq.Array(traderIDs), // trader_id = ANY(?)
        pq.Array(traderIDs), // WHERE trader_id = ANY(?)
        searchQuery.AmountFiat, searchQuery.AmountFiat, // лимиты сумм
    ).Scan(&finalCandidates).Error
    
    if err != nil {
        return nil, err
    }
    
    // Преобразование в доменные объекты
    bankDetails := make([]*domain.BankDetail, len(finalCandidates))
    for i, bankDetail := range finalCandidates {
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

func (r *DefaultBankDetailRepo) GetBankDetails(filter domain.GetBankDetailsFilter) ([]*domain.BankDetail, int64, error) {
	query := r.DB.Model(&models.BankDetailModel{})

	if filter.TraderID != nil {
		query = query.Where("trader_id = ?", *filter.TraderID)
	}
	if filter.BankCode != nil {
		query = query.Where("bank_code = ?", *filter.BankCode)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.PaymentSystem != nil {
		query = query.Where("payment_system = ?", *filter.PaymentSystem)
	}
	if filter.BankDetailID != nil {
		query = query.Where("id = ?", *filter.BankDetailID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count failed: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	var bankDetailModels []models.BankDetailModel
	if err := query.Find(&bankDetailModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find bank detail models: %w", err)
	}

	bankDetails := make([]*domain.BankDetail, len(bankDetailModels))
	for i, bankDetailModel := range bankDetailModels {
		bankDetails[i] = mappers.ToDomainBankDetail(&bankDetailModel)
	}

	return bankDetails, total, nil
}