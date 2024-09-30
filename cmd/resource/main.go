package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"stock-app/internal/api/timeseries"
	"stock-app/internal/cache"
	"stock-app/internal/repository"
	"stock-app/pkg/config"
	"stock-app/pkg/logger"
)

// Function to refresh data in database
func fetchLatestData(repo repository.StockRepo) {
	fmt.Println("Refreshing data...")
	tsFetcher := timeseries.NewTimeSeriesFetcher(config.AppConfig.TimeSeriesEndpoint, config.AppConfig.AlphaVantageAPIKey, config.AppConfig.SymbolList)

	if err := tsFetcher.FetchDailyData(repo); err != nil {
		fmt.Println("Failed to fetch latest data: ", err)
		os.Exit(1)
	}

	if err := tsFetcher.FetchIntradayData(repo); err != nil {
		fmt.Println("Failed to fetch latest data: ", err)
		os.Exit(1)
	}

	fmt.Println("Refreshed data in DB.")
}

// Function to build resources
func createTables(repo repository.StockRepo) {
	fmt.Println("Creating tables and indexing...")
	if err := repo.CreateTables(); err != nil {
		fmt.Println("Failed to create tables: ", err)
		os.Exit(1)
	}
	fmt.Println("Created tables in DB.")
	fetchLatestData(repo)
}

// Function to clean up resources
func cleanupCache(cache cache.StockCache) {
	fmt.Println("Cleaning up cache...")
	if err := cache.DeleteAll(); err != nil {
		fmt.Println("Failed to delete all cache data: ", err)
		os.Exit(1)
	}
	fmt.Println("Cleaned cache.")
}

func main() {
	// Define command-line flags
	createTableFlag := flag.Bool("create-tables", false, "Create tables")
	refreshFlag := flag.Bool("refresh", false, "Fetch latest data to DB")
	cleanupFlag := flag.Bool("cleanup", false, "Cleanup cache")

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
		fetchLatestData(repo)
	} else if *createTableFlag {
		createTables(repo)
	} else if *cleanupFlag {
		cleanupCache(cache)
	} else {
		fmt.Println("Usage: resource.go --refresh | --create-tables | --cleanup")
		os.Exit(1)
	}
}
