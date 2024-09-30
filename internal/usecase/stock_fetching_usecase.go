package usecase

import (
	"fmt"
	"time"

	"stock-app/internal/api/realtime"
	"stock-app/internal/cache"
	"stock-app/internal/entity"
	"stock-app/internal/repository"
	"stock-app/pkg/config"
	"stock-app/pkg/utils"
)

// StockFetchingUseCase defines the business logic related to stock data.
type StockFetchingUseCase struct {
	stockRepo       repository.StockRepo
	stockCache      cache.StockCache
	rtFetcher       *realtime.RealTimeFetcher
	latestQuoteData *entity.LatestQuoteData
}

func NewStockFetchingUseCase(
	stockRepo repository.StockRepo,
	stockCache cache.StockCache,
	rtFetcher *realtime.RealTimeFetcher,
	latestQuoteData *entity.LatestQuoteData,
) *StockFetchingUseCase {
	return &StockFetchingUseCase{
		stockRepo:       stockRepo,
		stockCache:      stockCache,
		rtFetcher:       rtFetcher,
		latestQuoteData: latestQuoteData,
	}
}

// FetchData update initial data to DB as service starts
func (sf *StockFetchingUseCase) FetchRealTimeData() error {
	fmt.Println("Fetching historical data ...")
	historicalData, err := sf.GetAllHistoricalData()
	if err != nil {
		return fmt.Errorf("failed to fetch historical data: %w", err)
	}
	fmt.Println("Successfully fetched historical data.")

	fmt.Println("Fetch and pre-populate latest data from cache to latestQuoteData...")
	if err := sf.PrePopulateLatestData(historicalData); err != nil {
		return fmt.Errorf("failed to fetch and pre-poluate latest data from cache: %w", err)
	}
	fmt.Println("Successfully fetched and pre-populated latest data to latestQuoteData.")

	// fmt.Println("Starting real-time updates...")
	// sf.rtFetcher.StartRealTimeUpdates(sf.latestQuoteData)
	// fmt.Println("Real-time updates started.")

	// fmt.Println("Start cron-job to Write data by minute...")
	// go sf.ScheduleDataWrite()

	return nil
}

func (sf *StockFetchingUseCase) GetAllHistoricalData() (map[string][]*entity.StockQuote, error) {
	startTime := time.Now().Add(-config.AppConfig.HistoricalDataDuration)
	endTime := time.Now()
	// Fetch historical data from cache
	historicalData, found := sf.stockCache.GetAll(startTime, endTime)
	if !found {
		fmt.Println("Cache is empty. Fetching historical data from DB (may need to refresh)...")
		historicalData, err := sf.stockRepo.GetAllHistoricalData(startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch historical data from DB: %w", err)
		}
		fmt.Printf("Fetched %d historical data from DB\n", len(historicalData))

		if err := sf.updateCache(historicalData); err != nil {
			return nil, err
		}
		fmt.Println("Successfully updated cache with historical data from DB.")
	} else {
		fmt.Println("Fetched historical data from cache.")
	}
	return historicalData, nil
}

func (sf *StockFetchingUseCase) PrePopulateLatestData(latestData map[string][]*entity.StockQuote) error {
	// Pre-populate latest data, preparing for real-time updates
	for symbol, quotes := range latestData {
		sf.latestQuoteData.Mu.Lock()
		fmt.Printf("Pre-populating latest data for symbol: %s with data: %v\n", symbol, quotes[len(quotes)-1])
		sf.latestQuoteData.StockData[symbol] = quotes[len(quotes)-1]
		sf.latestQuoteData.Mu.Unlock()
	}

	return nil
}

func (sf *StockFetchingUseCase) updateCache(latestData map[string][]*entity.StockQuote) error {
	var ttl time.Duration
	if utils.IsUSMarketOpen(time.Now()) {
		ttl = config.AppConfig.CacheShortTTL
	} else {
		ttl = config.AppConfig.CacheLongTTL
	}

	if err := sf.stockCache.SetAll(latestData, ttl); err != nil {
		return fmt.Errorf("failed to set all from list in cache: %w", err)
	}
	return nil
}

// ScheduleDataWrite schedules data write
func (sf *StockFetchingUseCase) ScheduleDataWrite() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	if utils.IsUSMarketOpen(time.Now()) {
		fmt.Println("US Market is open. Starting data Write cron-job...")
	} else {
		fmt.Println("US Market is closed. Exiting data Write cron-job...")
		return
	}

	for range ticker.C {
		if err := sf.writeDataToCache(); err != nil {
			fmt.Printf("Error during data Write: %v\n", err)
		}
		if err := sf.writeDataToDB(); err != nil {
			fmt.Printf("Error during data Write: %v\n", err)
		}
	}
}

func (sf *StockFetchingUseCase) writeDataToCache() error {
	sf.latestQuoteData.Mu.Lock()
	defer sf.latestQuoteData.Mu.Unlock()

	// Write data to cache
	if err := sf.stockCache.SetAllLatest(sf.latestQuoteData.StockData, config.AppConfig.CacheShortTTL); err != nil {
		return fmt.Errorf("error backing up data to cache: %v", err)
	}
	fmt.Printf("Successfully wrote data to cache\n")
	return nil
}

func (sf *StockFetchingUseCase) writeDataToDB() error {
	sf.latestQuoteData.Mu.Lock()
	defer sf.latestQuoteData.Mu.Unlock()

	for symbol, quote := range sf.latestQuoteData.StockData {
		timestampStr := quote.Timestamp.Format("2006-01-02 15:04:05")
		if err := sf.stockRepo.InsertIntradayData(
			symbol,
			timestampStr,
			fmt.Sprintf("%f", quote.OpenPrice),
			fmt.Sprintf("%f", quote.HighPrice),
			fmt.Sprintf("%f", quote.LowPrice),
			fmt.Sprintf("%f", quote.PrevClose),
			fmt.Sprintf("%f", quote.Volume),
		); err != nil {
			return fmt.Errorf("failed to write data for symbol %s: %w", symbol, err)
		}
	}
	fmt.Printf("Successfully wrote data to db\n")
	return nil
}
