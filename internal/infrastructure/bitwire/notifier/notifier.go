package notifier

import (
	"fmt"
	"log"
	"math"
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

        // Добавляем параметры в URL
        query := parsedURL.Query()
        query.Set("id", internalID)
        query.Set("status", status)
        if status == string(domain.StatusCompleted) {
            query.Set("usdRate", strconv.FormatFloat(reconciliationRate, 'f', 6, 64))
        }
        if reconciliationSum != 0 && reconciliationAmount != 0 && reconciliationRate != 0 {
            query.Set("reconciliationSum", strconv.FormatFloat(reconciliationSum, 'f', 6, 64))
            query.Set("reconciliationAmount", strconv.FormatFloat(reconciliationAmount, 'f', 6, 64))
            query.Set("reconciliationRate", strconv.FormatFloat(reconciliationRate, 'f', 6, 64))
        }
        parsedURL.RawQuery = query.Encode()
        targetURL := parsedURL.String()

        // Конфигурация ретраев
        maxAttempts := 3
        baseDelay := time.Second // Начальная задержка
        client := &http.Client{
            Timeout: 20 * time.Second, // Увеличенный таймаут
        }

        var lastError error
        for attempt := 0; attempt < maxAttempts; attempt++ {
            // Выполняем HTTP-запрос
            resp, err := client.Get(targetURL)
            if err == nil {
                // Закрываем тело при успешном ответе
                defer resp.Body.Close()
                
                // Обрабатываем статус ответа
                switch {
                case resp.StatusCode >= 200 && resp.StatusCode < 300:
                    log.Printf("callback success: sent to %s (attempt %d/%d)", 
                        targetURL, attempt+1, maxAttempts)
                    return
                    
                case resp.StatusCode >= 400 && resp.StatusCode < 500:
                    log.Printf("callback warning: client error %s for %s (attempt %d/%d)", 
                        resp.Status, targetURL, attempt+1, maxAttempts)
                    return
                    
                default:
                    lastError = fmt.Errorf("server error: %s", resp.Status)
                }
            } else {
                lastError = err
            }

            // Логируем неудачную попытку
            log.Printf("callback attempt failed: %s (attempt %d/%d): %v", 
                targetURL, attempt+1, maxAttempts, lastError)
            
            // Рассчитываем экспоненциальную задержку
            if attempt < maxAttempts-1 {
                delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
                time.Sleep(delay)
            }
        }

        // Финальная ошибка после всех попыток
        log.Printf("callback failed after %d attempts: %s: %v", 
            maxAttempts, targetURL, lastError)
    }()
}