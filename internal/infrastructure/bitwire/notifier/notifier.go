package notifier

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
)

func SendCallback(
	callbackUrl, 
	internalID, 
	status string,
	reconciliationSum,
	reconciliationAmount,
	reconciliationRate float64,
	) {
	go func() {
		// Парсим базовый URL
		parsedURL, err := url.Parse(callbackUrl)
		if err != nil {
			log.Printf("callback error: invalid URL '%s': %v", callbackUrl, err)
			return
		}

		switch status{
		case string(domain.StatusSucceed):
			status = "COMPLETED"
		case string(domain.StatusCanceled):
			status = "CANCELED"
		case string(domain.StatusDisputeCreated):
			status = "DISPUTE"
		case string(domain.StatusCreated):
			status = "PENDING"
		}

		query := parsedURL.Query()
		query.Set("id", internalID)
		query.Set("status", status)
		if reconciliationSum != 0 && reconciliationAmount != 0 && reconciliationRate != 0 {
			query.Set("reconciliationSum", strconv.FormatFloat(reconciliationSum, 'f', 6, 64))
			query.Set("reconciliationAmount", strconv.FormatFloat(reconciliationAmount, 'f', 6, 64))
			query.Set("reconciliationRate", strconv.FormatFloat(reconciliationRate, 'f', 6, 64))
		}
		parsedURL.RawQuery = query.Encode()

		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(parsedURL.String())
		if err != nil {
			log.Printf("callback error: request failed for %s: %v", parsedURL.String(), err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("callback warning: non-2xx response from %s: %s", parsedURL.String(), resp.Status)
		} else {
			log.Printf("callback success: sent to %s", parsedURL.String())
		}
	}()
}
