package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	walletResponse "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/response"
)

type HTTPWalletHandler struct {

}

func NewHTTPWalletHandler() (*HTTPWalletHandler, error) {
	return &HTTPWalletHandler{}, nil
}

func (uc *HTTPWalletHandler) Freeze(traderID, orderID string, amount float64) error {
	requestBodyBytes, err := json.Marshal(walletRequest.FreezeRequest{
		TraderID: traderID,
		OrderID: orderID,
		Amount: amount,
	})
	if err != nil {
		return err
	}

	response, err := http.Post("http://localhost:3000/wallets/freeze", "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}else {
		var errorResponse walletResponse.ErrorResponse
		if err := json.Unmarshal(responseBodyBytes, &errorResponse); err != nil {
			return err
		}
		return errors.New(errorResponse.Error)
	}
}

func (uc *HTTPWalletHandler) Release(traderID, orderID string, rewardPercent float64) error {
	requestBodyBytes, err := json.Marshal(walletRequest.ReleaseRequest{
		TraderID: traderID,
		OrderID: orderID,
		RewardPercent: rewardPercent,
	})
	if err != nil {
		return err
	}

	response, err := http.Post("http://localhost:3000/wallets/release", "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}else {
		var errorResponse walletResponse.ErrorResponse
		if err := json.Unmarshal(responseBodyBytes, &errorResponse); err != nil {
			return err
		}
		return errors.New(errorResponse.Error)
	}
}