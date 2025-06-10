package domain

import "errors"

var (
	ErrNoAvailableBankDetails = errors.New("no available bank details")
	ErrFreezeFailed = errors.New("freeze failed")
	ErrReleaseFailed = errors.New("release failed")
)