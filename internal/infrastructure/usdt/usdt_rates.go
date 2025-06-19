package usdt

import (
	"encoding/json"
	"io"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	UsdtRubRates = float64(0.0)
)

type RapiraItem struct {
	Price 	float64 `json:"price"`
	Amount  float64 `json:"amount"`
}

type RapiraResponse struct {
	Ask struct {
		Direction 	 string  	`json:"direction"`
		Symbol 		 string  	`json:"symbol"`
		MaxAmount	 float64 	`json:"max_amount"`
		MinAmount	 float64 	`json:"min_amount"`
		HighestPrice float64 	`json:"highest_price"`
		LowestPrice  float64 	`json:"lowest_price"`
		Items 		 []RapiraItem `json:"items"`
	}
}

func GET_USDT_RUB_RATES(ordersAmount int) (float64, error) {
	response, err := http.Get("https://api.rapira.net/market/exchange-plate-mini?symbol=USDT/RUB")
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		var rapiraResponse RapiraResponse
		if err := json.Unmarshal(responseBodyBytes, &rapiraResponse); err != nil {
			return 0, err
		}
		usdtAvgPrice := float64(0.0)
		for i := range ordersAmount {
			usdtAvgPrice += rapiraResponse.Ask.Items[i].Price
		}
		usdtAvgPrice /= float64(ordersAmount)
		UsdtRubRates = usdtAvgPrice
		return usdtAvgPrice, nil
	}

	return 0, status.Error(codes.Internal, "failed to count USDT average price in RUB")
}