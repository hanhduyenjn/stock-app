package timeseries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"stock-app/internal/entity"
	"stock-app/internal/repository"
)

// TimeSeriesFetcher manages real-time data from WebSocket API and external APIs.
type TimeSeriesFetcher struct {
	url     string
	symbols []string
}

// NewTimeSeriesFetcher creates a new instance of TimeSeriesFetcher.
func NewTimeSeriesFetcher(url string, apiToken string, symbols []string) *TimeSeriesFetcher {
	return &TimeSeriesFetcher{
		url:     url + "&apikey=" + apiToken,
		symbols: symbols,
	}
}

// FetchIntradayDataToDb fetches intraday data from the API and updates to DB
func (tf *TimeSeriesFetcher) FetchIntradayData(stockRepo repository.StockRepo) error {
	var wg sync.WaitGroup
	for _, symbol := range tf.symbols {
		wg.Add(1)
		go tf.fetchIntradayData(symbol, stockRepo, &wg)
	}
	wg.Wait()
	return nil
}

// fetchIntradayData fetches intraday data for a single symbol and updates to DB
func (tf *TimeSeriesFetcher) fetchIntradayData(symbol string, stockRepo repository.StockRepo, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Starting fetchIntradayData for symbol: %s\n", symbol)
	response, err := http.Get(tf.url + "&function=TIME_SERIES_INTRADAY&symbol=" + symbol + "&interval=1min")
	if err != nil {
		fmt.Printf("Error fetching intraday data for %s: %v\n", symbol, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Error response from API for %s: %s\n", symbol, response.Status)
		return
	}
	var apiResponse entity.TSIntradayResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		fmt.Printf("Error decoding JSON for %s: %v\n", symbol, err)
		return
	}

	fmt.Printf("Fetched data for symbol: %s, LastRefreshed: %s\n", symbol, apiResponse.MetaData.LastRefreshed)

	// Check if the latest timestamp matches the last refresh time
	lastRefresh := apiResponse.MetaData.LastRefreshed
	latestTimestamp, err := stockRepo.GetLatestIntradayDataTimestamp(symbol)
	if err != nil {
		fmt.Printf("Error fetching latest timestamp for %s: %v\n", symbol, err)
		return
	}

	fmt.Printf("Latest timestamp for symbol %s: %s\n", symbol, latestTimestamp)

	if (latestTimestamp != "" && latestTimestamp >= lastRefresh) {
		fmt.Printf("No new data for %s. Latest timestamp matches last refresh time.\n", symbol)
		return
	}

	// Iterate over Time Series and prepare data for insertion
	for timestamp, data := range apiResponse.TimeSeries {
		if timestamp <= latestTimestamp {
			fmt.Printf("Skipping data for symbol: %s, Timestamp: %s as it is before or equal to the latest timestamp from DB\n", symbol, timestamp)
			break
		}
		fmt.Printf("Inserting data for symbol: %s, Timestamp: %s\n", symbol, timestamp)
		err = stockRepo.InsertIntradayData(symbol, timestamp, data.Open, data.High, data.Low, data.Close, data.Volume)
		if err != nil {
			fmt.Printf("Error inserting intraday data for %s: %v\n", symbol, err)
		}
	}
	fmt.Printf("Completed fetchIntradayData for symbol: %s\n", symbol)
}

// FetchDailyDataToDB fetches historical data from the API and updates to DB
func (tf *TimeSeriesFetcher) FetchDailyData(stockRepo repository.StockRepo) error {
	var wg sync.WaitGroup
	for _, symbol := range tf.symbols {
		wg.Add(1)
		go tf.fetchDailyData(symbol, stockRepo, &wg)
	}
	wg.Wait()
	return nil
}

// fetchDailyData fetches daily data for a single symbol and updates to DB
func (tf *TimeSeriesFetcher) fetchDailyData(symbol string, stockRepo repository.StockRepo, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Starting fetchDailyData for symbol: %s\n", symbol)
	response, err := http.Get(tf.url + "&function=TIME_SERIES_DAILY&symbol=" + symbol)
	if err != nil {
		fmt.Printf("Error fetching daily data for %s: %v\n", symbol, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Error response from API for %s: %s\n", symbol, response.Status)
		return
	}

	var apiResponse entity.TSDailyResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		fmt.Printf("Error decoding JSON for %s: %v\n", symbol, err)
		return
	}

	fmt.Printf("Fetched data for symbol: %s, LastRefreshed: %s\n", symbol, apiResponse.MetaData.LastRefreshed)

	// Check if the latest date matches the last refresh date
	lastRefresh := apiResponse.MetaData.LastRefreshed
	latestDate, err := stockRepo.GetLatestDailyDataDate(symbol)
	if err != nil {
		fmt.Printf("Error fetching latest date for %s: %v\n", symbol, err)
		return
	}

	fmt.Printf("Latest date for symbol %s: %s\n", symbol, latestDate)

	if (latestDate != "" && latestDate >= lastRefresh) {
		fmt.Printf("No new data for %s. Latest date matches last refresh date.\n", symbol)
		return
	}

	// Iterate over Time Series and prepare data for insertion
	for date, data := range apiResponse.TimeSeries {
		if date <= latestDate {
			fmt.Printf("Skipping data for symbol: %s, Date: %s as it is before or equal to the latest date from DB\n", symbol, date)
			continue
		}
		fmt.Printf("Inserting data for symbol: %s, Date: %s\n", symbol, date)
		err = stockRepo.InsertDailyData(symbol, date, data.Open, data.High, data.Low, data.Close, data.Volume)
		if err != nil {
			fmt.Printf("Error inserting daily data for %s: %v\n", symbol, err)
		}
	}
	fmt.Printf("Completed fetchDailyData for symbol: %s\n", symbol)
}
