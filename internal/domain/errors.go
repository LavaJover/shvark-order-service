package domain

import "errors"

var (
	ErrNoAvailableBankDetails = errors.New("no available bank details")
	ErrFreezeFailed = errors.New("freeze failed")
	ErrReleaseFailed = errors.New("release failed")
	ErrOpenDisputeFailed = errors.New("failed to open dispute")
	ErrResolveDisputeFailed = errors.New("failed to resolve dispute")
	ErrCancelOrder = errors.New("failed to cancel order")
)