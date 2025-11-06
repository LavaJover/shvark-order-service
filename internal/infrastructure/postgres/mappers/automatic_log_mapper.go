package mappers

import (
    "strings"
    
    "github.com/LavaJover/shvark-order-service/internal/domain"
    "github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
)

// ToDomainAutomaticLog конвертирует модель в доменный объект
func ToDomainAutomaticLog(model *models.AutomaticLogModel) *domain.AutomaticLog {
    if model == nil {
        return nil
    }
    
    methods := []string{}
    if model.Methods != "" {
        methods = strings.Split(model.Methods, ",")
    }
    
    return &domain.AutomaticLog{
        ID:             model.ID,
        DeviceID:       model.DeviceID,
        TraderID:       model.TraderID,
        OrderID:        model.OrderID,
        Amount:         model.Amount,
        PaymentSystem:  model.PaymentSystem,
        Direction:      model.Direction,
        Methods:        methods,
        ReceivedAt:     model.ReceivedAt,
        Text:           model.Text,
        Action:         model.Action,
        Success:        model.Success,
        OrdersFound:    model.OrdersFound,
        ErrorMessage:   model.ErrorMessage,
        ProcessingTime: model.ProcessingTime,
        BankName:       model.BankName,
        CardNumber:     model.CardNumber,
        CreatedAt:      model.CreatedAt,
    }
}

// ToModelAutomaticLog конвертирует доменный объект в модель
func ToModelAutomaticLog(log *domain.AutomaticLog) *models.AutomaticLogModel {
    if log == nil {
        return nil
    }
    
    methods := ""
    if len(log.Methods) > 0 {
        methods = strings.Join(log.Methods, ",")
    }
    
    return &models.AutomaticLogModel{
        ID:             log.ID,
        DeviceID:       log.DeviceID,
        TraderID:       log.TraderID,
        OrderID:        log.OrderID,
        Amount:         log.Amount,
        PaymentSystem:  log.PaymentSystem,
        Direction:      log.Direction,
        Methods:        methods,
        ReceivedAt:     log.ReceivedAt,
        Text:           log.Text,
        Action:         log.Action,
        Success:        log.Success,
        OrdersFound:    log.OrdersFound,
        ErrorMessage:   log.ErrorMessage,
        ProcessingTime: log.ProcessingTime,
        BankName:       log.BankName,
        CardNumber:     log.CardNumber,
        CreatedAt:      log.CreatedAt,
    }
}
