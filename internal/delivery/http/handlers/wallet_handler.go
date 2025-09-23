package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	response, err := http.Post(fmt.Sprintf("http://%s/wallets/freeze", h.Address), "application/json", bytes.NewBuffer(requestBodyBytes))
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

func (h *HTTPWalletHandler) Release(releaseRequest walletRequest.ReleaseRequest) error {
	requestBodyBytes, err := json.Marshal(releaseRequest)
	if err != nil {
		return err
	}

	response, err := http.Post(fmt.Sprintf("http://%s/wallets/release", h.Address), "application/json", bytes.NewBuffer(requestBodyBytes))
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
	response, err := http.Get(fmt.Sprintf("http://%s/wallets/%s/balance", h.Address, traderID))
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

// GetTraderBalancesBatch - метод для получения балансов конкретных трейдеров
func (h *HTTPWalletHandler) GetTraderBalancesBatch(traderIDs []string) (map[string]float64, error) {
	if len(traderIDs) == 0 {
		return map[string]float64{}, nil
	}

	// Простой GET запрос
	url := fmt.Sprintf("http://%s/wallets/balances/batch?traderIds=%s", 
		h.Address, strings.Join(traderIDs, ","))
	
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", response.StatusCode, string(body))
	}

	var result struct {
		Success  bool              `json:"success"`
		Balances []walletResponse.BalanceResponse `json:"balances"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, errors.New("request failed")
	}

	balances := make(map[string]float64)
	for _, b := range result.Balances {
		balances[b.TraderID] = b.Balance
	}

	return balances, nil
}