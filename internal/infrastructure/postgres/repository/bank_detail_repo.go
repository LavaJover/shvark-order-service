package repository

import (
	"fmt"
	"log"
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

// Вспомогательная функция для безопасного маскирования карты
func maskCardNumber(cardNumber string) string {
    if len(cardNumber) >= 4 {
        return "***" + cardNumber[len(cardNumber)-4:]
    } else if cardNumber != "" {
        return cardNumber
    }
    return "N/A"
}

func (r *DefaultBankDetailRepo) FindSuitableBankDetails(searchQuery *domain.SuitablleBankDetailsQuery) ([]*domain.BankDetail, error) {
    log.Printf("=== START FindSuitableBankDetails ===")
    log.Printf("SearchQuery: AmountFiat=%.2f, PaymentSystem=%s, Currency=%s, BankCode=%s, NspkCode=%s",
        searchQuery.AmountFiat, searchQuery.PaymentSystem, searchQuery.Currency, 
        searchQuery.BankCode, searchQuery.NspkCode)
    
    baseCandidates, err := r.findBaseCandidates(searchQuery)
    if err != nil {
        log.Printf("ERROR: findBaseCandidates failed: %v", err)
        return nil, err
    }
    
    log.Printf("Stage 1 (findBaseCandidates): Found %d candidates", len(baseCandidates))
    
    if len(baseCandidates) == 0 {
        log.Printf("WARNING: No base candidates found. Returning empty result.")
        return []*domain.BankDetail{}, nil
    }

    finalCandidates, err := r.applyDynamicConstraintsOptimized(baseCandidates, searchQuery)
    if err != nil {
        log.Printf("ERROR: applyDynamicConstraintsOptimized failed: %v", err)
        return nil, err
    }

    log.Printf("Stage 2 (applyDynamicConstraints): Final %d candidates", len(finalCandidates))
    log.Printf("=== END FindSuitableBankDetails ===\n")

    return finalCandidates, nil
}

// Этап 1 остается без изменений
// Этап 1: Находим базовые кандидаты по статическим параметрам
func (r *DefaultBankDetailRepo) findBaseCandidates(searchQuery *domain.SuitablleBankDetailsQuery) ([]models.BankDetailModel, error) {
    log.Printf("\n--- Stage 1: findBaseCandidates ---")
    
    var baseCandidates []models.BankDetailModel
    
    query := r.DB.Model(&models.BankDetailModel{}).
        Where("enabled = ?", true).
        Where("min_amount <= ? AND max_amount >= ?", searchQuery.AmountFiat, searchQuery.AmountFiat).
        Where("payment_system = ?", searchQuery.PaymentSystem).
        Where("currency = ?", searchQuery.Currency).
        Where("deleted_at IS NULL")
    
    log.Printf("Base filters: enabled=true, amount range includes %.2f, payment_system=%s, currency=%s",
        searchQuery.AmountFiat, searchQuery.PaymentSystem, searchQuery.Currency)
    
    if searchQuery.BankCode != "" {
        query = query.Where("bank_code = ?", searchQuery.BankCode)
        log.Printf("Additional filter: bank_code=%s", searchQuery.BankCode)
    }
    
    if searchQuery.NspkCode != "" {
        query = query.Where("nspk_code = ?", searchQuery.NspkCode)
        log.Printf("Additional filter: nspk_code=%s", searchQuery.NspkCode)
    }
    
    err := query.Find(&baseCandidates).Error
    
    if err != nil {
        log.Printf("ERROR: Database query failed: %v", err)
        return nil, err
    }
    
    log.Printf("\nFound %d base candidates:", len(baseCandidates))
    for i, candidate := range baseCandidates {
        log.Printf("  [%d] TraderID=%s, Card=%s, Amount=[%.2f-%.2f], PaymentSystem=%s, Currency=%s, BankCode=%s, Enabled=%v, MaxSimultaneous=%d",
            i+1,
            candidate.TraderID,
            maskCardNumber(candidate.CardNumber),
            candidate.MinAmount,
            candidate.MaxAmount,
            candidate.PaymentSystem,
            candidate.Currency,
            candidate.BankCode,
            candidate.Enabled,
            candidate.MaxOrdersSimultaneosly,
        )
    }
    
    return baseCandidates, err
}

// КАРДИНАЛЬНО ОПТИМИЗИРОВАННЫЙ этап 2 с использованием денормализации
// Этап 2: Применяем динамические ограничения
func (r *DefaultBankDetailRepo) applyDynamicConstraintsOptimized(baseCandidates []models.BankDetailModel, searchQuery *domain.SuitablleBankDetailsQuery) ([]*domain.BankDetail, error) {
    log.Printf("\n--- Stage 2: applyDynamicConstraintsOptimized ---")
    
    if len(baseCandidates) == 0 {
        return []*domain.BankDetail{}, nil
    }
    
    // Получаем IDs реквизитов (НЕ trader_id!)
    bankDetailIDs := make([]string, len(baseCandidates))
    for i, candidate := range baseCandidates {
        bankDetailIDs[i] = candidate.ID
    }
    
    log.Printf("Checking dynamic constraints for %d bank details", len(bankDetailIDs))
    
    now := time.Now()
    startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
    
    log.Printf("Time ranges: Now=%s, StartOfDay=%s, StartOfMonth=%s",
        now.Format("2006-01-02 15:04:05"),
        startOfDay.Format("2006-01-02 15:04:05"),
        startOfMonth.Format("2006-01-02 15:04:05"))
    
    // ИСПРАВЛЕННЫЙ SQL - группируем по bank_details_id
    sqlQuery := `
        WITH bank_detail_stats AS (
            SELECT 
                bank_details_id::text as bank_details_id_text,
                -- Количество активных заказов
                SUM(CASE WHEN status = $1 THEN 1 ELSE 0 END)::int as pending_count,
                -- Статистика за день
                SUM(CASE WHEN status = ANY($2::text[]) AND created_at >= $3 THEN 1 ELSE 0 END)::int as day_count,
                SUM(CASE WHEN status = ANY($4::text[]) AND created_at >= $5 THEN amount_fiat ELSE 0 END)::float as day_amount,
                -- Статистика за месяц
                SUM(CASE WHEN status = ANY($6::text[]) AND created_at >= $7 THEN 1 ELSE 0 END)::int as month_count,
                SUM(CASE WHEN status = ANY($8::text[]) AND created_at >= $9 THEN amount_fiat ELSE 0 END)::float as month_amount,
                -- Время последнего завершенного заказа
                MAX(CASE WHEN status = $10 THEN created_at END) as last_completed_time,
                -- Проверка на дублирующие заказы
                SUM(CASE WHEN status = $11 AND amount_fiat = $12 THEN 1 ELSE 0 END)::int as duplicate_count
            FROM order_models 
            WHERE bank_details_id::text = ANY($13::text[])
            GROUP BY bank_details_id
        ),
        debug_stats AS (
            SELECT 
                bd.id::text as bank_detail_id,
                bd.trader_id::text as trader_id,
                bd.card_number,
                bd.max_orders_simultaneosly,
                bd.max_quantity_day,
                bd.max_amount_day,
                bd.max_quantity_month,
                bd.max_amount_month,
                bd.delay,
                COALESCE(bds.pending_count, 0) as pending_count,
                COALESCE(bds.day_count, 0) as day_count,
                COALESCE(bds.day_amount, 0) as day_amount,
                COALESCE(bds.month_count, 0) as month_count,
                COALESCE(bds.month_amount, 0) as month_amount,
                bds.last_completed_time,
                COALESCE(bds.duplicate_count, 0) as duplicate_count,
                CASE WHEN COALESCE(bds.pending_count, 0) >= bd.max_orders_simultaneosly THEN 'max_simultaneous' ELSE NULL END as reason_1,
                CASE WHEN COALESCE(bds.day_count, 0) + 1 > bd.max_quantity_day THEN 'max_day_count' ELSE NULL END as reason_2,
                CASE WHEN COALESCE(bds.day_amount, 0) + $15 > bd.max_amount_day THEN 'max_day_amount' ELSE NULL END as reason_3,
                CASE WHEN COALESCE(bds.month_count, 0) + 1 > bd.max_quantity_month THEN 'max_month_count' ELSE NULL END as reason_4,
                CASE WHEN COALESCE(bds.month_amount, 0) + $16 > bd.max_amount_month THEN 'max_month_amount' ELSE NULL END as reason_5,
                CASE WHEN bds.last_completed_time IS NOT NULL AND bds.last_completed_time > NOW() - (bd.delay / 1000000000.0) * INTERVAL '1 SECOND' THEN 'delay_not_passed' ELSE NULL END as reason_6,
                CASE WHEN COALESCE(bds.duplicate_count, 0) > 0 THEN 'duplicate_order' ELSE NULL END as reason_7
            FROM bank_detail_models bd
            LEFT JOIN bank_detail_stats bds ON bd.id::text = bds.bank_details_id_text
            WHERE bd.id::text = ANY($14::text[])
              AND bd.enabled = true
              AND bd.deleted_at IS NULL
        )
        SELECT * FROM debug_stats
    `
    
    type DebugStats struct {
        BankDetailID          string
        TraderID              string
        CardNumber            string
        MaxOrdersSimultaneous int32
        MaxQuantityDay        int32
        MaxAmountDay          float64
        MaxQuantityMonth      int32
        MaxAmountMonth        float64
        Delay                 time.Duration
        PendingCount          int
        DayCount              int
        DayAmount             float64
        MonthCount            int
        MonthAmount           float64
        LastCompletedTime     *time.Time
        DuplicateCount        int
        Reason1               *string
        Reason2               *string
        Reason3               *string
        Reason4               *string
        Reason5               *string
        Reason6               *string
        Reason7               *string
    }
    
    var debugStats []DebugStats
    
    pendingCompletedStatuses := []string{string(domain.StatusPending), string(domain.StatusCompleted)}
    
    err := r.DB.Raw(sqlQuery,
        string(domain.StatusPending),           // $1
        pq.Array(pendingCompletedStatuses),     // $2
        startOfDay,                             // $3
        pq.Array(pendingCompletedStatuses),     // $4
        startOfDay,                             // $5
        pq.Array(pendingCompletedStatuses),     // $6
        startOfMonth,                           // $7
        pq.Array(pendingCompletedStatuses),     // $8
        startOfMonth,                           // $9
        string(domain.StatusCompleted),         // $10
        string(domain.StatusPending),           // $11
        searchQuery.AmountFiat,                 // $12
        pq.Array(bankDetailIDs),               // $13 - bank_details_id
        pq.Array(bankDetailIDs),               // $14 - bank_details_id
        searchQuery.AmountFiat,                 // $15
        searchQuery.AmountFiat,                 // $16
    ).Scan(&debugStats).Error
    
    if err != nil {
        log.Printf("ERROR: Debug stats query failed: %v", err)
        return nil, fmt.Errorf("failed to get debug stats: %w", err)
    }
    
    // Логируем детальную статистику
    log.Printf("\nDetailed stats for %d candidates:", len(debugStats))
    for i, stat := range debugStats {
        reasons := []string{}
        if stat.Reason1 != nil { reasons = append(reasons, *stat.Reason1) }
        if stat.Reason2 != nil { reasons = append(reasons, *stat.Reason2) }
        if stat.Reason3 != nil { reasons = append(reasons, *stat.Reason3) }
        if stat.Reason4 != nil { reasons = append(reasons, *stat.Reason4) }
        if stat.Reason5 != nil { reasons = append(reasons, *stat.Reason5) }
        if stat.Reason6 != nil { reasons = append(reasons, *stat.Reason6) }
        if stat.Reason7 != nil { reasons = append(reasons, *stat.Reason7) }
        
        status := "✓ PASSED"
        if len(reasons) > 0 {
            status = fmt.Sprintf("✗ REJECTED: %v", reasons)
        }
        
        lastCompleted := "NEVER"
        if stat.LastCompletedTime != nil {
            lastCompleted = stat.LastCompletedTime.Format("15:04:05")
        }
        
        log.Printf("  [%d] %s", i+1, status)
        log.Printf("      BankDetailID=%s, TraderID=%s, Card=%s", 
            stat.BankDetailID, stat.TraderID, maskCardNumber(stat.CardNumber))
        log.Printf("      Pending: %d/%d, DayCount: %d+1/%d, DayAmount: %.2f+%.2f/%.2f",
            stat.PendingCount, stat.MaxOrdersSimultaneous,
            stat.DayCount, stat.MaxQuantityDay,
            stat.DayAmount, searchQuery.AmountFiat, stat.MaxAmountDay)
        log.Printf("      MonthCount: %d+1/%d, MonthAmount: %.2f+%.2f/%.2f",
            stat.MonthCount, stat.MaxQuantityMonth,
            stat.MonthAmount, searchQuery.AmountFiat, stat.MaxAmountMonth)
        log.Printf("      LastCompleted: %s, Delay: %v, Duplicates: %d",
            lastCompleted, stat.Delay, stat.DuplicateCount)
    }
    
    // Финальный запрос с правильной группировкой
    finalQuery := `
        WITH bank_detail_stats AS (
            SELECT 
                bank_details_id::text as bank_details_id_text,
                SUM(CASE WHEN status = $1 THEN 1 ELSE 0 END)::int as pending_count,
                SUM(CASE WHEN status = ANY($2::text[]) AND created_at >= $3 THEN 1 ELSE 0 END)::int as day_count,
                SUM(CASE WHEN status = ANY($4::text[]) AND created_at >= $5 THEN amount_fiat ELSE 0 END)::float as day_amount,
                SUM(CASE WHEN status = ANY($6::text[]) AND created_at >= $7 THEN 1 ELSE 0 END)::int as month_count,
                SUM(CASE WHEN status = ANY($8::text[]) AND created_at >= $9 THEN amount_fiat ELSE 0 END)::float as month_amount,
                MAX(CASE WHEN status = $10 THEN created_at END) as last_completed_time,
                SUM(CASE WHEN status = $11 AND amount_fiat = $12 THEN 1 ELSE 0 END)::int as duplicate_count
            FROM order_models 
            WHERE bank_details_id::text = ANY($13::text[])
            GROUP BY bank_details_id
        )
        SELECT bd.* 
        FROM bank_detail_models bd
        LEFT JOIN bank_detail_stats bds ON bd.id::text = bds.bank_details_id_text
        WHERE bd.id::text = ANY($14::text[])
          AND bd.enabled = true
          AND bd.deleted_at IS NULL
          AND COALESCE(bds.pending_count, 0) < bd.max_orders_simultaneosly
          AND COALESCE(bds.day_count, 0) + 1 <= bd.max_quantity_day
          AND COALESCE(bds.day_amount, 0) + $15 <= bd.max_amount_day
          AND COALESCE(bds.month_count, 0) + 1 <= bd.max_quantity_month
          AND COALESCE(bds.month_amount, 0) + $16 <= bd.max_amount_month
          AND (bds.last_completed_time IS NULL OR bds.last_completed_time <= NOW() - (bd.delay / 1000000000.0) * INTERVAL '1 SECOND')
          AND COALESCE(bds.duplicate_count, 0) = 0
    `
    
    var finalCandidates []models.BankDetailModel
    
    err = r.DB.Raw(finalQuery,
        string(domain.StatusPending),
        pq.Array(pendingCompletedStatuses),
        startOfDay,
        pq.Array(pendingCompletedStatuses),
        startOfDay,
        pq.Array(pendingCompletedStatuses),
        startOfMonth,
        pq.Array(pendingCompletedStatuses),
        startOfMonth,
        string(domain.StatusCompleted),
        string(domain.StatusPending),
        searchQuery.AmountFiat,
        pq.Array(bankDetailIDs),              // $13 - используем bank_details_id
        pq.Array(bankDetailIDs),              // $14 - используем bank_details_id
        searchQuery.AmountFiat,
        searchQuery.AmountFiat,
    ).Scan(&finalCandidates).Error
    
    if err != nil {
        log.Printf("ERROR: Final query failed: %v", err)
        return nil, fmt.Errorf("failed to apply dynamic constraints: %w", err)
    }
    
    log.Printf("\nFinal result: %d candidates passed all checks", len(finalCandidates))
    
    bankDetails := make([]*domain.BankDetail, len(finalCandidates))
    for i, bankDetail := range finalCandidates {
        bankDetails[i] = mappers.ToDomainBankDetail(&bankDetail)
        log.Printf("  Final [%d] BankDetailID=%s, TraderID=%s, Card=%s", 
            i+1, bankDetail.ID, bankDetail.TraderID, maskCardNumber(bankDetail.CardNumber))
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