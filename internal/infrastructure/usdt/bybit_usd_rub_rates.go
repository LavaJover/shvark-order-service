package usdt

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type OrderbookResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		Symbol    string     `json:"s"`
		Bids      [][]string `json:"b"`
		Asks      [][]string `json:"a"`
		Timestamp int64      `json:"ts"`
		UpdateID  int64      `json:"u"`
	} `json:"result"`
	Time int64 `json:"time"`
}

type TickerResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		Category string `json:"category"`
		List     []struct {
			Symbol      string `json:"symbol"`
			LastPrice   string `json:"lastPrice"`
			Bid1Price   string `json:"bid1Price"`
			Ask1Price   string `json:"ask1Price"`
			Volume24h   string `json:"volume24h"`
			Price24hPcnt string `json:"price24hPcnt"`
		} `json:"list"`
	} `json:"result"`
}

type InstrumentsResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		Category string `json:"category"`
		List     []struct {
			Symbol     string `json:"symbol"`
			BaseCoin   string `json:"baseCoin"`
			QuoteCoin  string `json:"quoteCoin"`
			Status     string `json:"status"`
		} `json:"list"`
	} `json:"result"`
}

const (
	BYBIT_API_BASE = "https://api.bybit.com"
	ORDERBOOK_ENDPOINT = "/v5/market/orderbook"
	TICKER_ENDPOINT = "/v5/market/tickers"
	INSTRUMENTS_ENDPOINT = "/v5/market/instruments-info"
)

// HTTP клиент с таймаутом
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Функция для безопасного HTTP запроса
func makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Добавляем заголовки
	req.Header.Set("User-Agent", "RUB-USD-Calculator/1.0")
	req.Header.Set("Accept", "application/json")

	return httpClient.Do(req)
}

// Функция для поиска RUB пар в списке инструментов
func findRubPairs() ([]string, error) {
	url := fmt.Sprintf("%s%s?category=spot", BYBIT_API_BASE, INSTRUMENTS_ENDPOINT)

	resp, err := makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	var instruments InstrumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&instruments); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	if instruments.RetCode != 0 {
		return nil, fmt.Errorf("API ошибка: %s", instruments.RetMsg)
	}

	var rubPairs []string
	for _, instrument := range instruments.Result.List {
		symbol := strings.ToUpper(instrument.Symbol)
		if strings.Contains(symbol, "RUB") && instrument.Status == "Trading" {
			rubPairs = append(rubPairs, symbol)
		}
	}

	return rubPairs, nil
}

// Функция для получения стакана (orderbook)
func getOrderbook(symbol string, category string) (*OrderbookResponse, error) {
	url := fmt.Sprintf("%s%s?category=%s&symbol=%s&limit=10", 
		BYBIT_API_BASE, ORDERBOOK_ENDPOINT, category, symbol)

	resp, err := makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	var orderbook OrderbookResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderbook); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	if orderbook.RetCode != 0 {
		return nil, fmt.Errorf("API ошибка: %s", orderbook.RetMsg)
	}

	return &orderbook, nil
}

// Функция для получения тикера
func getTicker(symbol string, category string) (*TickerResponse, error) {
	url := fmt.Sprintf("%s%s?category=%s&symbol=%s", 
		BYBIT_API_BASE, TICKER_ENDPOINT, category, symbol)

	resp, err := makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	var ticker TickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	if ticker.RetCode != 0 {
		return nil, fmt.Errorf("API ошибка: %s", ticker.RetMsg)
	}

	return &ticker, nil
}

// Функция для вычисления среднего курса из bid ордеров (зелёный стакан)
func calculateAverageFromBids(bids [][]string, count int) (float64, error) {
	if len(bids) == 0 {
		return 0, fmt.Errorf("нет bid ордеров")
	}

	// Ограничиваем количество ордеров от 4 до 6, как указано в задании
	if count < 4 {
		count = 4
	}
	if count > 6 || count > len(bids) {
		count = len(bids)
		if count > 6 {
			count = 6
		}
	}

	var sum float64
	var validCount int

	fmt.Printf("\n📊 Анализ bid ордеров (зелёный стакан):\n")
	for i := 0; i < count && i < len(bids); i++ {
		if len(bids[i]) >= 2 {
			price, err := strconv.ParseFloat(bids[i][0], 64)
			if err != nil {
				log.Printf("Ошибка парсинга цены %s: %v", bids[i][0], err)
				continue
			}
			volume, _ := strconv.ParseFloat(bids[i][1], 64)
			fmt.Printf("   %d. Цена: %s | Объём: %.6f\n", i+1, bids[i][0], volume)
			sum += price
			validCount++
		}
	}

	if validCount == 0 {
		return 0, fmt.Errorf("не удалось распарсить ни одну цену")
	}

	average := sum / float64(validCount)
	fmt.Printf("\n💹 Средняя цена по %d bid ордерам: %.8f\n", validCount, average)
	return average, nil
}

// Функция для расчёта кросс-курса USD/RUB через BTC
func calculateCrossRate() {
	fmt.Println("\n🔄 Расчёт кросс-курса USD/RUB через BTC...")

	// Получаем BTC/USDT
	btcUsdt, err := getOrderbook("BTCUSDT", "spot")
	if err != nil {
		fmt.Printf("❌ Не удалось получить BTC/USDT: %v\n", err)
		return
	}

	if len(btcUsdt.Result.Bids) >= 4 {
		avgBtcUsd, err := calculateAverageFromBids(btcUsdt.Result.Bids, 5)
		if err != nil {
			fmt.Printf("❌ Ошибка расчёта BTC/USD: %v\n", err)
			return
		}

		fmt.Printf("\n📈 BTC/USD средняя цена: $%.2f\n", avgBtcUsd)
		fmt.Printf("💡 Для полного расчёта RUB/USD нужен курс BTC/RUB из внешнего источника\n")
		fmt.Printf("   Формула: USD/RUB = (BTC/USD) / (BTC/RUB)\n")

		rubBtc, err := getOrderbook("RUBBTC", "spot")
		if err != nil {
			fmt.Printf("❌ Не удалось получить RUB/BTC: %v\n", err)
			return
		}

		// if len(rubBtc.Result.Bids) >= 4 {
		// 	avgBtcRub, err := calculateAverageFromBids(rubBtc.Result.Bids, 5)
		// 	if err != nil {
		// 		fmt.Printf("❌ Ошибка расчёта : %v\n", err)
		// 		return
		// 	}
		// }

		fmt.Printf("\n📊 Примерные расчёты (на основе USD/RUB ≈ 83):\n")
		fmt.Printf("   BTC/RUB ≈ ₽%.0f\n", rubBtc)
		// fmt.Printf("   USD/RUB ≈ ₽%.2f\n", estimatedUsdRub)
	}
}

// Основная функция для получения курса RUB/USD
func GetRubToUsdRate() {
	fmt.Println("🔍 Поиск RUB торговых пар на Bybit...")

	// Сначала попробуем найти все RUB пары
	rubPairs, err := findRubPairs()
	if err != nil {
		fmt.Printf("⚠️  Не удалось получить список инструментов: %v\n", err)
		fmt.Println("Переходим к проверке стандартных пар...")
	} else {
		fmt.Printf("✅ Найдено RUB пар: %d\n", len(rubPairs))
		for _, pair := range rubPairs {
			fmt.Printf("   - %s\n", pair)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// Проверяем наиболее вероятные символы
	testSymbols := []string{
		"USDTRUB", "RUBUSDT", "USDCRUB", "RUBUSDC", 
		"EURUSDT", // для кросс-курса
	}

	var found bool

	for _, symbol := range testSymbols {
		fmt.Printf("\n🔍 Проверяем %s...\n", symbol)

		// Пробуем получить orderbook
		orderbook, err := getOrderbook(symbol, "spot")
		if err != nil {
			fmt.Printf("  ❌ Orderbook: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ Orderbook найден!\n")
		fmt.Printf("     Bid ордеров: %d\n", len(orderbook.Result.Bids))
		fmt.Printf("     Ask ордеров: %d\n", len(orderbook.Result.Asks))

		if len(orderbook.Result.Bids) >= 4 {
			avgRate, err := calculateAverageFromBids(orderbook.Result.Bids, 5)
			if err != nil {
				fmt.Printf("  ❌ Ошибка вычисления: %v\n", err)
				continue
			}

			fmt.Printf("\n🎯 УСПЕХ! Найден курс для %s\n", symbol)
			fmt.Printf("📈 Средний курс по bid ордерам: %.8f\n", avgRate)

			// Интерпретируем результат
			if strings.Contains(symbol, "USDT") && strings.Contains(symbol, "RUB") {
				if strings.HasPrefix(symbol, "RUB") {
					fmt.Printf("💱 Это означает: 1 RUB = %.8f USDT\n", avgRate)
					fmt.Printf("💱 Или: 1 USD ≈ %.2f RUB\n", 1.0/avgRate)
				} else {
					fmt.Printf("💱 Это означает: 1 USDT = %.2f RUB\n", avgRate)
					fmt.Printf("💱 Или: 1 RUB = %.8f USD\n", 1.0/avgRate)
				}
			}

			found = true
			break
		}
	}

	if !found {
		fmt.Println("\n❌ Прямые RUB/USD пары не найдены в orderbook")
		fmt.Println("\n🔄 Попробуем альтернативный метод...")
		calculateCrossRate()

		fmt.Println("\n💡 Рекомендации:")
		fmt.Println("   1. Используйте P2P торговлю Bybit для RUB/USDT")
		fmt.Println("   2. Получите курс с внешних источников (ЦБ РФ, Forex)")
		fmt.Println("   3. Используйте кросс-курс через EUR или другие валюты")
	}

	fmt.Printf("\n⏰ Запрос завершён: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
