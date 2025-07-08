package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	walletResponse "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/response"
)

type HTTPWalletHandler struct {
	Address string
}

func NewHTTPWalletHandler(address string) (*HTTPWalletHandler, error) {
	return &HTTPWalletHandler{
		Address: address,
	}, nil
}

func (h *HTTPWalletHandler) Freeze(traderID, orderID string, amount float64) error {
	requestBodyBytes, err := json.Marshal(walletRequest.FreezeRequest{
		TraderID: traderID,
		OrderID: orderID,
		Amount: amount,
	})
	if err != nil {
		return err
	}

	response, err := http.Post(fmt.Sprintf("%s/wallets/freeze", h.Address), "application/json", bytes.NewBuffer(requestBodyBytes))
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

func (h *HTTPWalletHandler) Release(traderID, merchantID, orderID string, rewardPercent, platformFee float64) error {
	requestBodyBytes, err := json.Marshal(walletRequest.ReleaseRequest{
		TraderID: traderID,
		OrderID: orderID,
		RewardPercent: rewardPercent,
		MerchantID: merchantID,
		PlatformFee: platformFee,
	})
	if err != nil {
		return err
	}

	response, err := http.Post(fmt.Sprintf("%s/wallets/release", h.Address), "application/json", bytes.NewBuffer(requestBodyBytes))
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

func (h *HTTPWalletHandler) GetTraderBalance(traderID string) (float64, error) {
	response, err := http.Get(fmt.Sprintf("%s/wallets/%s/balance", h.Address, traderID))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		var balanceResponse walletResponse.BalanceResponse
		if err := json.Unmarshal(responseBodyBytes, &balanceResponse); err != nil {
			return 0, err
		}
		return balanceResponse.Balance, nil
	}
	var errorResponse walletResponse.ErrorResponse
	if err := json.Unmarshal(responseBodyBytes, &errorResponse); err != nil {
		return 0, err
	}
	return 0, errors.New(errorResponse.Error)
}