package usecase

import (
	"fmt"
	"time"

	"stock-app/internal/api/realtime"
	"stock-app/internal/api/timeseries"
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
	tsFetcher       *timeseries.TimeSeriesFetcher
	latestQuoteData *entity.LatestQuoteData
}

func NewStockFetchingUseCase(
	stockRepo repository.StockRepo,
	stockCache cache.StockCache,
	rtFetcher *realtime.RealTimeFetcher,
	tsFetcher *timeseries.TimeSeriesFetcher,
	latestQuoteData *entity.LatestQuoteData,
) *StockFetchingUseCase {
	return &StockFetchingUseCase{
		stockRepo:       stockRepo,
		stockCache:      stockCache,
		rtFetcher:       rtFetcher,
		tsFetcher:       tsFetcher,
		latestQuoteData: latestQuoteData,
	}
}

// FetchData update initial data to DB as service starts
func (sf *StockFetchingUseCase) FetchRealTimeData() error {
	fmt.Println("Fetch and pre-populate data from cache...")
	if err := sf.PrePopulateLatestData(); err != nil {
		return fmt.Errorf("failed to fetch and pre-poluate latest data from cache: %w", err)
	}

	fmt.Println("Starting real-time updates...")
	sf.rtFetcher.StartRealTimeUpdates(sf.latestQuoteData)
	fmt.Println("Real-time updates started.")

	fmt.Println("Start cron-job to Write data by minute...")
	go sf.ScheduleDataWrite()

	return nil
}

func (sf *StockFetchingUseCase) PrePopulateLatestData() error {
	// Fetch latest data from cache
	latestData, found := sf.stockCache.GetAll()
	if !found {
		fmt.Println("Cache is empty. Fetch latest data from API...")
		if err := sf.FetchLatestData(); err != nil {
			return fmt.Errorf("failed to fetch latest data: %w", err)
		}
		cacheData, _ := sf.stockCache.GetAll()
		latestData = append(latestData, cacheData...)
	}

	fmt.Println("Cache is pre-populated with latest data.")
	for _, quote := range latestData {
		sf.latestQuoteData.Mu.Lock()
		sf.latestQuoteData.StockData[quote.Symbol] = quote
		sf.latestQuoteData.Mu.Unlock()
		fmt.Printf("Fetched data for symbol %s: %+v\n", quote.Symbol, quote)
	}
	return nil
}

func (sf *StockFetchingUseCase) FetchLatestData() error {
	fmt.Println("Fetching intraday data from API...")
	if err := sf.tsFetcher.FetchIntradayData(sf.stockRepo); err != nil {
		return fmt.Errorf("failed to fetch intraday data: %w", err)
	}
	fmt.Println("Intraday data fetched successfully.")

	fmt.Println("Fetching daily data from API...")
	if err := sf.tsFetcher.FetchDailyData(sf.stockRepo); err != nil {
		return fmt.Errorf("failed to fetch daily data: %w", err)
	}
	fmt.Println("Daily data fetched successfully.")

	latestData, err := sf.stockRepo.GetAllLatestIntradayData()
	if err != nil {
		return fmt.Errorf("failed to fetch latest data from DB: %w", err)
	}
	fmt.Printf("Fetched %d latest data from DB\n", len(latestData))

	if err := sf.updateCache(latestData); err != nil {
		return err
	}
	fmt.Println("Successfully updated cache with latest data.")

	return nil
}

func (sf *StockFetchingUseCase) updateCache(latestData []*entity.StockQuote) error {
	if err := sf.stockCache.DeleteAll(); err != nil {
		return fmt.Errorf("failed to delete all from cache: %w", err)
	}

	var ttl time.Duration
	if utils.IsUSMarketOpen(time.Now()) {
		ttl = config.AppConfig.CacheShortTTL
	} else {
		ttl = config.AppConfig.CacheLongTTL
	}

	if err := sf.stockCache.SetAllFromList(latestData, ttl); err != nil {
		return fmt.Errorf("failed to set all from list in cache: %w", err)
	}
	return nil
}

// ScheduleDataWrite schedules data write
func (sf *StockFetchingUseCase) ScheduleDataWrite() {
	ticker := time.NewTicker(time.Second*2)
	defer ticker.Stop()

	if utils.IsUSMarketOpen(time.Now()) {
		fmt.Println("US Market is open. Starting data Write cron-job...")
	} else {
		fmt.Println("US Market is closed. Exiting data Write cron-job...")
		return
	}

	for range ticker.C {
		if err := sf.writeData(); err != nil {
			fmt.Printf("Error during data Write: %v\n", err)
		}
	}
}

func (sf *StockFetchingUseCase) writeData() error {
	// Write data to cache, then to DB (write-back cache)
	if err := sf.stockCache.SetAll(sf.latestQuoteData, config.AppConfig.CacheShortTTL); err != nil {
		return fmt.Errorf("error backing up data to cache: %v", err)

	}
	latestData, found := sf.stockCache.GetAll()
	if !found {
		return fmt.Errorf("error fetching data from cache")
	}

	for _, quote := range latestData {
		timestampStr := quote.Timestamp.Format("2006-01-02 15:04:05")
		if err := sf.stockRepo.InsertIntradayData(
			quote.Symbol,
			timestampStr,
			fmt.Sprintf("%f", quote.OpenPrice),
			fmt.Sprintf("%f", quote.HighPrice),
			fmt.Sprintf("%f", quote.LowPrice),
			fmt.Sprintf("%f", quote.PrevClose),
			fmt.Sprintf("%f", quote.Volume),
		); err != nil {
			return fmt.Errorf("failed to Write data for symbol %s: %w", quote.Symbol, err)
		}
	}
	
	fmt.Printf("Successfully Write data to DB\n")
	return nil
}