package repository

import (
	"gorm.io/gorm"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

type MerchantStoreRepository struct {
    db *gorm.DB
}

func NewMerchantStoreRepository(db *gorm.DB) *MerchantStoreRepository {
    return &MerchantStoreRepository{db: db}
}

func (r *MerchantStoreRepository) CreateMerchantStore(store *domain.MerchantStore) error {
    model := &models.MerchantStoreModel{
        ID:            store.ID,
        MerchantID:    store.MerchantID,
        Name:          store.Name,
        PlatformFee:   store.PlatformFee,
        IsActive:      store.IsActive,
        DealsDuration: store.DealsDuration,
        Description:   store.Description,
        Category:      store.Category,
        MaxDailyDeals: store.MaxDailyDeals,
        MinDealAmount: store.MinDealAmount,
        MaxDealAmount: store.MaxDealAmount,
        Currency:      store.Currency,
        CreatedAt:     store.CreatedAt,
        UpdatedAt:     store.UpdatedAt,
    }
    
    return r.db.Create(model).Error
}

func (r *MerchantStoreRepository) UpdateMerchantStore(store *domain.MerchantStore) error {
    model := &models.MerchantStoreModel{
        ID:            store.ID,
        MerchantID:    store.MerchantID,
        Name:          store.Name,
        PlatformFee:   store.PlatformFee,
        IsActive:      store.IsActive,
        DealsDuration: store.DealsDuration,
        Description:   store.Description,
        Category:      store.Category,
        MaxDailyDeals: store.MaxDailyDeals,
        MinDealAmount: store.MinDealAmount,
        MaxDealAmount: store.MaxDealAmount,
        Currency:      store.Currency,
        UpdatedAt:     store.UpdatedAt,
    }
    
    return r.db.Save(model).Error
}

func (r *MerchantStoreRepository) DeleteMerchantStore(id string) error {
    return r.db.Delete(&models.MerchantStoreModel{}, "id = ?", id).Error
}

func (r *MerchantStoreRepository) GetMerchantStoreByID(id string) (*domain.MerchantStore, error) {
    var model models.MerchantStoreModel
    if err := r.db.First(&model, "id = ?", id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, err
    }
    
    return r.toDomain(&model), nil
}

func (r *MerchantStoreRepository) GetMerchantStores(page, limit int32) ([]*domain.MerchantStore, error) {
    var storeModels []*models.MerchantStoreModel
    
    offset := (page - 1) * limit
    query := r.db.Model(&models.MerchantStoreModel{})
    
    if err := query.Offset(int(offset)).Limit(int(limit)).Find(&storeModels).Error; err != nil {
        return nil, err
    }
    
    return r.toDomainList(storeModels), nil
}

func (r *MerchantStoreRepository) GetStoresByMerchantID(merchantID string) ([]*domain.MerchantStore, error) {
    var models []*models.MerchantStoreModel
    
    if err := r.db.Where("merchant_id = ?", merchantID).Find(&models).Error; err != nil {
        return nil, err
    }
    
    return r.toDomainList(models), nil
}

func (r *MerchantStoreRepository) GetActiveStoresByMerchantID(merchantID string) ([]*domain.MerchantStore, error) {
    var models []*models.MerchantStoreModel
    
    if err := r.db.Where("merchant_id = ? AND is_active = ?", merchantID, true).Find(&models).Error; err != nil {
        return nil, err
    }
    
    return r.toDomainList(models), nil
}

func (r *MerchantStoreRepository) toDomain(model *models.MerchantStoreModel) *domain.MerchantStore {
    return &domain.MerchantStore{
        ID:            model.ID,
        MerchantID:    model.MerchantID,
        Name:          model.Name,
        PlatformFee:   model.PlatformFee,
        IsActive:      model.IsActive,
        DealsDuration: model.DealsDuration,
        Description:   model.Description,
        Category:      model.Category,
        MaxDailyDeals: model.MaxDailyDeals,
        MinDealAmount: model.MinDealAmount,
        MaxDealAmount: model.MaxDealAmount,
        Currency:      model.Currency,
        CreatedAt:     model.CreatedAt,
        UpdatedAt:     model.UpdatedAt,
    }
}

func (r *MerchantStoreRepository) toDomainList(models []*models.MerchantStoreModel) []*domain.MerchantStore {
    stores := make([]*domain.MerchantStore, len(models))
    for i, model := range models {
        stores[i] = r.toDomain(model)
    }
    return stores
}