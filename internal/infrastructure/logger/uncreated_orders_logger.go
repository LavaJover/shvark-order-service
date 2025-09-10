package logger

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/usecase"
)

type UncreatedOrdersLogger interface {
	LogUncreatedOrder(event *domain.UncreatedOrder) error
}

type DefaultUncreatedOrdersLogger struct {
	loggerUsecase usecase.UncreatedOrderUsecase
}

func NewDefaultUncreatedOrdersLogger(loggerUsecase usecase.UncreatedOrderUsecase) *DefaultUncreatedOrdersLogger {
	return &DefaultUncreatedOrdersLogger{
		loggerUsecase: loggerUsecase,
	}
}

func (l *DefaultUncreatedOrdersLogger) LogUncreatedOrder(event *domain.UncreatedOrder) error {
	if err := l.loggerUsecase.LogEvent(event); err != nil {
		return err
	}
	return nil
}
