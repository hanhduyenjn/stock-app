package latestquote

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"stock-app/internal/cache"
	"stock-app/internal/entity"
	"stock-app/pkg/config"
)

// LatestQuoteFetcher manages real-time data from WebSocket API and external APIs.
type LatestQuoteFetcher struct {
	url     string
	symbols []string
}

// NewLatestQuoteFetcher creates a new instance of LatestQuoteFetcher.
func NewLatestQuoteFetcher(url string, apiToken string, symbols []string) *LatestQuoteFetcher {
	return &LatestQuoteFetcher{
		url:     url + "?token=" + apiToken,
		symbols: symbols,
	}
}

// FetchToCache fetches latest quote data from the external API and updates the cache.
func (qf *LatestQuoteFetcher) FetchToCache(stockCache cache.StockCache) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorChannel := make(chan error, len(qf.symbols))

	fetchData := func(symbol string) {
		defer wg.Done()
		url := fmt.Sprintf("%s&symbol=%s", qf.url, symbol)

		for {
			fmt.Printf("Fetching data for symbol %s from URL: %s\n", symbol, url)

			resp, err := http.Get(url)
			if err != nil {
				errorChannel <- fmt.Errorf("failed to fetch data for symbol %s: %w", symbol, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				retryAfter := resp.Header.Get("Retry-After")
				if retryAfter != "" {
					duration, err := time.ParseDuration(retryAfter + "s")
					if err != nil {
						duration = time.Minute
					}
					fmt.Printf("Rate limit exceeded for symbol %s. Retrying after %v...\n", symbol, duration)
					time.Sleep(duration)
					continue
				} else {
					fmt.Printf("Rate limit exceeded for symbol %s. Retrying after 1 minute...\n", symbol)
					time.Sleep(time.Minute)
					continue
				}
			}

			if resp.StatusCode != http.StatusOK {
				errorChannel <- fmt.Errorf("non-OK HTTP status for symbol %s: %s", symbol, resp.Status)
				return
			}

			var stockQuote entity.StockQuote
			if err := json.NewDecoder(resp.Body).Decode(&stockQuote); err != nil {
				errorChannel <- fmt.Errorf("failed to decode data for symbol %s: %w", symbol, err)
				return
			}
			stockQuote.Symbol = symbol

			fmt.Printf("Fetched data for symbol %s: %+v\n", symbol, stockQuote)

			mu.Lock()
			stockCache.Set(symbol, &stockQuote, config.AppConfig.CacheShortTTL)
			mu.Unlock()

			break
		}
	}

	for _, symbol := range qf.symbols {
		wg.Add(1)
		go fetchData(symbol)
	}

	go func() {
		wg.Wait()
		close(errorChannel)
	}()

	var err error
	for e := range errorChannel {
		if err == nil {
			err = e
		} else {
			err = fmt.Errorf("%w; %v", err, e)
		}
	}

	if err != nil {
		fmt.Printf("Errors encountered: %v\n", err)
		return err
	}

	fmt.Println("Successfully fetched and updated stock data")
	return nil
}
