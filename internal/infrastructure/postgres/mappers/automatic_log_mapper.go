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
    
    // Создаем доменный объект с безопасным разыменованием указателей
    domainLog := &domain.AutomaticLog{
        ID:             model.ID,
        DeviceID:       model.DeviceID,
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
    
    // Безопасное разыменование TraderID
    if model.TraderID != nil {
        domainLog.TraderID = *model.TraderID
    } else {
        domainLog.TraderID = "" // или какое-то значение по умолчанию
    }
    
    // Безопасное разыменование OrderID
    if model.OrderID != nil {
        domainLog.OrderID = *model.OrderID
    } else {
        domainLog.OrderID = "" // или какое-то значение по умолчанию
    }
    
    return domainLog
}

// ToModelAutomaticLog конвертирует доменный объект в модель
func ToModelAutomaticLog(log *domain.AutomaticLog) *models.AutomaticLogModel {
    model := &models.AutomaticLogModel{
        ID:             log.ID,
        DeviceID:       log.DeviceID,
        Amount:         log.Amount,
        PaymentSystem:  log.PaymentSystem,
        Direction:      log.Direction,
        Methods:        strings.Join(log.Methods, ","),
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
    
    // Обрабатываем nullable поля
    if log.TraderID != "" && log.TraderID != "00000000-0000-0000-0000-000000000000" {
        model.TraderID = &log.TraderID
    }
    
    if log.OrderID != "" && log.OrderID != "00000000-0000-0000-0000-000000000000" {
        model.OrderID = &log.OrderID
    }
    
    return model
}