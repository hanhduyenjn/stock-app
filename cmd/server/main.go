package main

import (
	"database/sql"
	"sync"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"stock-app/internal/api/realtime"
	"stock-app/internal/cache"
	"stock-app/internal/entity"
	"stock-app/internal/handler"
	"stock-app/internal/repository"
	"stock-app/internal/usecase"
	"stock-app/pkg/config"
	"stock-app/pkg/logger"
)

func main() {
	// Load configuration
	config.LoadConfig()
	log := logger.NewLogger()

	// Initialize Gin Router
	router := gin.Default()

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
	rtStockData := &entity.LatestQuoteData{
		StockData: make(map[string]*entity.StockQuote), // Initialize the map
		Mu:        sync.RWMutex{},                      // Initialize the mutex
	}

	repo := repository.NewStockRepo(dbConn)
	cache := cache.NewStockCache(config.AppConfig.CacheClient)
	stockServingUseCase := usecase.NewStockServingUseCase(repo, cache, rtStockData)

	rtFetcher := realtime.NewRealTimeFetcher(config.AppConfig.RealTimeTradesEndpoint, config.AppConfig.FinnhubAPIKey, config.AppConfig.SymbolList)
	stockFetchingUseCase := usecase.NewStockFetchingUseCase(repo, cache, rtFetcher, rtStockData)

	// Fetch data in real-time
	if err := stockFetchingUseCase.FetchRealTimeData(); err != nil {
		log.Fatal("Failed to fetch initial data: ", err)
	}

	stockHandler := handler.NewStockHandler(stockServingUseCase)

	// Stock Management endpoints
    stock := router.Group("/stocks")
    {
        stock.GET("", stockHandler.GetAllQuotes)
        stock.GET("/quote", stockHandler.GetQuote) // The handler will receive `symbol` and `start` with `end` as query parameters
        // stock.GET("/trade", stockHandler.GetTrades) // Similar to above, `symbol` and `range` are query parameters
        // stock.GET("/profile", stockHandler.GetCompanyProfile) // `symbol` can be a query parameter
        // stock.GET("/financials", stockHandler.GetFinancials) // `symbol` can be a query parameter
    }
    

	// Start the server on the configured port
	port := config.AppConfig.ServerPort
	log.Printf("Starting HTTP server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
