package notifier

import (
    "fmt"
    "log"
    "math"
    "net/http"
    "net/url"
    "strconv"
    "time"
)

func SendCallback(
    callbackUrl, 
    internalID, 
    status string,
    reconciliationSum,
    reconciliationAmount,
    reconciliationRate float64,
) error {
    // Парсим базовый URL
    parsedURL, err := url.Parse(callbackUrl)
    if err != nil {
        return fmt.Errorf("invalid callback URL '%s': %w", callbackUrl, err)
    }

    // Добавляем параметры в URL
    query := parsedURL.Query()
    query.Set("id", internalID)
    query.Set("status", status)
    if reconciliationSum != 0 && reconciliationAmount != 0 && reconciliationRate != 0 {
        query.Set("reconciliationSum", strconv.FormatFloat(reconciliationSum, 'f', 6, 64))
        query.Set("reconciliationAmount", strconv.FormatFloat(reconciliationAmount, 'f', 6, 64))
        query.Set("reconciliationRate", strconv.FormatFloat(reconciliationRate, 'f', 6, 64))
    }
    parsedURL.RawQuery = query.Encode()
    targetURL := parsedURL.String()

    // Конфигурация ретраев
    maxAttempts := 3
    baseDelay := time.Second
    client := &http.Client{
        Timeout: 20 * time.Second,
    }

    var lastError error
    for attempt := 0; attempt < maxAttempts; attempt++ {
        // Выполняем HTTP-запрос
        resp, err := client.Get(targetURL)
        if err == nil {
            defer resp.Body.Close()
            
            // Обрабатываем статус ответа
            switch {
            case resp.StatusCode >= 200 && resp.StatusCode < 300:
                log.Printf("callback success: sent to %s (attempt %d/%d)", 
                    targetURL, attempt+1, maxAttempts)
                return nil // Успешно отправлено
                
            case resp.StatusCode >= 400 && resp.StatusCode < 500:
                // Клиентская ошибка - не ретраим
                err := fmt.Errorf("client error: %s", resp.Status)
                log.Printf("callback client error: %s for %s (attempt %d/%d)", 
                    resp.Status, targetURL, attempt+1, maxAttempts)
                return err
                
            default:
                lastError = fmt.Errorf("server error: %s", resp.Status)
            }
        } else {
            lastError = err
        }

        // Логируем неудачную попытку
        log.Printf("callback attempt failed: %s (attempt %d/%d): %v", 
            targetURL, attempt+1, maxAttempts, lastError)
        
        // Рассчитываем экспоненциальную задержку между попытками
        if attempt < maxAttempts-1 {
            delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
            time.Sleep(delay)
        }
    }

    // Возвращаем ошибку после всех неудачных попыток
    finalError := fmt.Errorf("callback failed after %d attempts to %s: %w", 
        maxAttempts, targetURL, lastError)
    log.Printf("callback final failure: %v", finalError)
    return finalError
}

// SendCallbackAsync - асинхронная версия для обратной совместимости
func SendCallbackAsync(
    callbackUrl, 
    internalID, 
    status string,
    reconciliationSum,
    reconciliationAmount,
    reconciliationRate float64,
) {
    go func() {
        if err := SendCallback(callbackUrl, internalID, status, reconciliationSum, reconciliationAmount, reconciliationRate); err != nil {
            log.Printf("async callback error: %v", err)
        }
    }()
}