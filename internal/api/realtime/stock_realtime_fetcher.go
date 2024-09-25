package realtime

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"

	"stock-app/internal/entity"
	"stock-app/pkg/utils"
)

// RealTimeFetcher manages real-time data from WebSocket API.
type RealTimeFetcher struct {
	wsURL   string
	symbols []string
}

// NewRealTimeFetcher creates a new instance of the real-time RealTimeFetcher.
func NewRealTimeFetcher(wsURL, apiToken string, symbols []string) *RealTimeFetcher {
	return &RealTimeFetcher{
		wsURL:   wsURL + "?token=" + apiToken,
		symbols: symbols}
}

// StartRealTimeUpdates starts fetching real-time updates and updating the in-memory storage.
func (h *RealTimeFetcher) StartRealTimeUpdates(latestQuoteData *entity.LatestQuoteData) {
	go func() {
		// Connect to WebSocket
		fmt.Printf("Connecting to WebSocket at URL: %s\n", h.wsURL)
		conn, _, err := websocket.DefaultDialer.Dial(h.wsURL, nil)
		if err != nil {
			fmt.Printf("Failed to connect to WebSocket: %v\n", err)
			return
		}
		defer conn.Close()
		fmt.Println("WebSocket connection established.")

		// Subscribe to stock symbols
		for _, symbol := range h.symbols {
			msg := map[string]interface{}{"type": "subscribe", "symbol": symbol}
			fmt.Printf("Subscribing to symbol: %s\n", symbol)
			if err := conn.WriteJSON(msg); err != nil {
				fmt.Printf("Failed to send subscription message for %s: %v\n", symbol, err)
				return
			}
		}

		for {
			var response map[string]interface{}
			err := conn.ReadJSON(&response)
			if err != nil {
				fmt.Printf("Error reading WebSocket data: %v\n", err)
				continue
			}

			fmt.Printf("Received response from WebSocket: %v\n", response)

			if response["type"] == "trade" {
				trades, ok := response["data"].([]interface{})
				if !ok {
					fmt.Printf("Unexpected data format: %v\n", response["data"])
					continue
				}

				fmt.Printf("Processing trades: %v\n", trades)

				for _, trade := range trades {
					tradeData, ok := trade.(map[string]interface{})
					if !ok {
						fmt.Printf("Unexpected trade format: %v\n", trade)
						continue
					}

					symbol := tradeData["s"].(string)
					price := tradeData["p"].(float64)
					timestamp := int64(tradeData["t"].(float64))
					volume := tradeData["v"].(float64)

					fmt.Printf("Trade received for symbol %s: Price = %.2f, Volume = %.2f, Timestamp = %d\n", symbol, price, volume, timestamp)

					// Fetch historical data for calculations
					latestQuoteData.Mu.RLock()
					prevQuote, exists := latestQuoteData.StockData[symbol]
					latestQuoteData.Mu.RUnlock()

					if !exists {
						fmt.Printf("No previous data for symbol %s\n", symbol)
						continue // Skip updating this symbol as historical data is missing
					}

					fmt.Printf("Previous data for %s: %+v\n", symbol, prevQuote)

					// Calculate changes based on historical data
					change := price - prevQuote.Price
					changePercentage := (change / prevQuote.Price) * 100
					highPrice := utils.Max(price, prevQuote.HighPrice)
					lowPrice := utils.Min(price, prevQuote.LowPrice)
					currentVolume := prevQuote.Volume + volume

					// Create StockQuote with updated values
					stockQuote := &entity.StockQuote{
						Symbol:           symbol,
						Price:            price,
						Change:           change,
						ChangePercentage: changePercentage,
						HighPrice:        highPrice,
						LowPrice:         lowPrice,
						OpenPrice:        prevQuote.OpenPrice,
						PrevClose:        prevQuote.Price,
						Volume:           currentVolume,
						Timestamp:        time.Unix(0, timestamp*int64(time.Millisecond)),
					}

					fmt.Printf("Updated stock data for %s: %+v\n", symbol, stockQuote)

					// Update real-time data in-memory
					latestQuoteData.Mu.Lock()
					latestQuoteData.StockData[symbol] = stockQuote
					latestQuoteData.Mu.Unlock()

					fmt.Printf("Real-time data updated for symbol %s\n", symbol)
				}
			}
		}
	}()
}
