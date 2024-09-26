package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"

	"stock-app/internal/api/realtime"
	"stock-app/internal/api/timeseries"
	"stock-app/internal/cache"
	"stock-app/internal/entity"
	"stock-app/internal/repository"
	"stock-app/internal/usecase"
	"stock-app/pkg/config"
	"stock-app/pkg/logger"
)

// Function to refresh data in database
func refreshResources(repo repository.StockRepo, cache cache.StockCache) {
	fmt.Println("Refreshing resources...")
	rtFetcher := realtime.NewRealTimeFetcher(config.AppConfig.RealTimeTradesEndpoint, config.AppConfig.FinnhubAPIKey, config.AppConfig.SymbolList)
	tsFetcher := timeseries.NewTimeSeriesFetcher(config.AppConfig.TimeSeriesEndpoint, config.AppConfig.AlphaVantageAPIKey, config.AppConfig.SymbolList)

	rtStockData := &entity.LatestQuoteData{
		StockData: make(map[string]*entity.StockQuote), // Initialize the map
		Mu:        sync.RWMutex{},                      // Initialize the mutex
	}
	stockFetchingUseCase := usecase.NewStockFetchingUseCase(repo, cache, rtFetcher, tsFetcher, rtStockData)
	if err := stockFetchingUseCase.FetchLatestData(); err != nil {
		fmt.Println("Failed to fetch latest data: ", err)
		os.Exit(1)
	}
	fmt.Println("Refreshed resources.")
}

// Function to build resources
func createTables(repo repository.StockRepo, cache cache.StockCache) {
	fmt.Println("Building resources...")
	if err := repo.CreateTables(); err != nil {
		fmt.Println("Failed to create tables: ", err)
		os.Exit(1)
	}
	fmt.Println("Built resources.")
	refreshResources(repo, cache)
}

// Function to clean up resources
func cleanupResources(cache cache.StockCache) {
	fmt.Println("Cleaning up resources...")
	if err := cache.DeleteAll(); err != nil {
		fmt.Println("Failed to delete all cache data: ", err)
		os.Exit(1)
	}
	fmt.Println("Cleaned resources.")
}

func main() {
	// Define command-line flags
	refreshFlag := flag.Bool("refresh", false, "Refresh resources")
	buildFlag := flag.Bool("build", false, "Build resources")
	cleanupFlag := flag.Bool("cleanup", false, "Cleanup resources")

	// Parse the command-line flags
	flag.Parse()

	// Load configuration
	config.LoadConfig()
	log := logger.NewLogger()

	// Initialize database connection
	dbConn, err := sql.Open("postgres", config.AppConfig.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to the database: ", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Fatal("Failed to close the database connection: ", err)
		}
	}()

	// Initialize dependencies
	repo := repository.NewStockRepo(dbConn)
	cache := cache.NewStockCache(config.AppConfig.CacheClient)

	// Check which flag was set and call the corresponding function
	if *refreshFlag {
		refreshResources(repo, cache)
	} else if *buildFlag {
		createTables(repo, cache)
	} else if *cleanupFlag {
		cleanupResources(cache)
	} else {
		fmt.Println("Usage: resource.go --refresh | --build | --cleanup")
		os.Exit(1)
	}
}
