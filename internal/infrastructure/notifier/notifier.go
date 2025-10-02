package notifier

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func SendCallback(callbackURL string, payload CallbackPayload) {
	go func() {
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal callback: %v\n", err)
			return
		}

		req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Failed to create callback request: %v\n", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		// Header с HMAC сигнатурой

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Callback failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Callback sent to %s\n", callbackURL)
		}else {
			log.Printf("Callback returned status %d", resp.StatusCode)
		}
	}()
}