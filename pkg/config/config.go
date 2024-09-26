package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"stock-app/pkg/utils"

	"github.com/joho/godotenv"
)

// Config holds the configuration values loaded from environment variables or .env file
type Config struct {
    AlphaVantageAPIKey     string
    TimeSeriesEndpoint     string
    FinnhubAPIKey          string
    QuoteEndpoint          string
    RealTimeTradesEndpoint string
    SymbolList             []string
    DatabaseURL            string
    CacheClient            string
    CacheShortTTL          time.Duration
    CacheLongTTL           time.Duration
    ServerPort             string
    LogLevel               string
}

// AppConfig is the global configuration instance
var AppConfig Config

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() {
    // Load .env file if it exists
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found or failed to load .env file")
    }

    // Initialize AppConfig with environment variables
    AppConfig = Config{
        AlphaVantageAPIKey:     getEnv("ALPHA_VANTAGE_API_KEY", ""),
        TimeSeriesEndpoint:     getEnv("TIMESERIES_ENDPOINT", ""),
        FinnhubAPIKey:          getEnv("FINHUBB_API_KEY", ""),
        QuoteEndpoint:          getEnv("QUOTE_ENDPOINT", ""),
        RealTimeTradesEndpoint: getEnv("REAL_TIME_TRADES_ENDPOINT", ""),
        SymbolList:             getSymbolList(getEnv("SYMBOL_LIST", "AAPL,TSLA,GOOGL,AMZN,MSFT")),
        DatabaseURL:            getDBConnectionString(),
        CacheClient:            getRedisConnectionString(),
        CacheShortTTL:          getCacheTTL("CACHE_SHORT_TTL", 10),
        CacheLongTTL:           getCacheTTL("CACHE_LONG_TTL", 24000),
        ServerPort:             getEnv("SERVER_PORT", "8080"),
        LogLevel:               getEnv("LOG_LEVEL", "debug"),
    }
}

// getEnv retrieves an environment variable or returns a default value if not set
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

// getDBConnectionString constructs a database connection string from environment variables
func getDBConnectionString() string {
    username := getEnv("DB_USERNAME", "postgres")
    password := getEnv("DB_PASSWORD", "mysecretpassword")
    host := getEnv("DB_HOST", "localhost")
    port := getEnv("DB_PORT", "5432")
    dbname := getEnv("DB_NAME", "stockdatabase")

    return "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=disable"
}

// getRedisConnectionString constructs the Redis connection string
func getRedisConnectionString() string {
    host := getEnv("REDIS_HOST", "localhost")
    port := getEnv("REDIS_PORT", "6379")
    return host + ":" + port
}

// getSymbolList parses the SYMBOL_LIST environment variable into a slice of strings
func getSymbolList(symbols string) []string {
    if symbols == "" {
        return []string{}
    }
    return strings.Split(symbols, ",")
}

// getCacheTTL retrieves and converts the cache TTL value from the environment variable
func getCacheTTL(key string, defaultTTL int) time.Duration {
    return time.Duration(utils.ToInt(getEnv(key, strconv.Itoa(defaultTTL)))) * time.Second
}
