package notifier

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

func SendCallback(
	callbackUrl, 
	internalID, 
	status string,
	) {
	go func() {
		// Парсим базовый URL
		parsedURL, err := url.Parse(callbackUrl)
		if err != nil {
			log.Printf("callback error: invalid URL '%s': %v", callbackUrl, err)
			return
		}

		query := parsedURL.Query()
		query.Set("id", internalID)
		query.Set("status", status)
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
