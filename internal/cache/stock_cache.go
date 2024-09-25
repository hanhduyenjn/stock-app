package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/go-redis/redis/v8"

    "stock-app/internal/entity"
)

var ctx = context.Background()

// StockCache defines the interface for caching stock data.
type StockCache interface {
    Get(symbol string) (*entity.StockQuote, bool)
    Set(symbol string, stock *entity.StockQuote, expiration time.Duration)
    GetAll() ([]*entity.StockQuote, bool)
    SetAll(latestQuoteData *entity.LatestQuoteData, expiration time.Duration) error
    SetAllFromList(stocks []*entity.StockQuote, expiration time.Duration) error
    DeleteAll() error
}

// RedisStockCache is a Redis-backed cache for stock data.
type RedisStockCache struct {
    client     *redis.Client
}

// NewStockCache creates a new RedisStockCache instance with a specified expiration time.
func NewStockCache(redisAddr string) StockCache {
    rdb := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    return &RedisStockCache{
        client:     rdb,
    }
}

// Get retrieves a stock from the cache by symbol.
func (c *RedisStockCache) Get(symbol string) (*entity.StockQuote, bool) {
    val, err := c.client.Get(ctx, symbol).Result()
    if err == redis.Nil {
        return nil, false // Cache miss
    }
    if err != nil {
        return nil, false // Redis error
    }

    var stock entity.StockQuote
    if err := json.Unmarshal([]byte(val), &stock); err != nil {
        return nil, false // JSON unmarshalling error
    }

    return &stock, true
}

// GetAll retrieves all stocks from the cache.
func (c *RedisStockCache) GetAll() ([]*entity.StockQuote, bool) {
    var stocks []*entity.StockQuote
    keys, err := c.client.Keys(ctx, "*").Result()
    if err != nil {
        return nil, false // Redis error
    }

    for _, key := range keys {
        stock, found := c.Get(key)
        if found {
            stocks = append(stocks, stock)
        }
    }

    return stocks, len(stocks) > 0
}

// Set stores a stock in the cache with an optional expiration time.
func (c *RedisStockCache) Set(symbol string, stock *entity.StockQuote, expiration time.Duration) {
    c.setCache(symbol, stock, expiration)
}

// SetAll stores multiple stocks in the cache with an optional expiration time.
func (c *RedisStockCache) SetAll(latestQuoteData *entity.LatestQuoteData, expiration time.Duration) error {
    var mu sync.Mutex
    for _, stock := range latestQuoteData.StockData {
        mu.Lock()
        c.setCache(stock.Symbol, stock, expiration)
        mu.Unlock()
    }
    return nil
}

// SetAllFromList stores multiple stocks in the cache from a list, with an optional expiration time.
func (c *RedisStockCache) SetAllFromList(stocks []*entity.StockQuote, expiration time.Duration) error {
    for _, stock := range stocks {
        c.setCache(stock.Symbol, stock, expiration)
    }
    return nil
}

// setCache is a helper function to handle setting data in the cache with expiration.
func (c *RedisStockCache) setCache(symbol string, stock *entity.StockQuote, expiration time.Duration) {
    data, err := json.Marshal(stock)
    if err != nil {
        fmt.Printf("Failed to marshal stock data for %s: %v\n", symbol, err)
        return // Handle JSON marshalling error, log if needed
    }

    status := c.client.Set(ctx, symbol, data, expiration)
    if err := status.Err(); err != nil {
        fmt.Printf("Failed to cache stock %s: %v\n", symbol, err)
    } else {
        fmt.Printf("Successfully cached stock %s\n", symbol)
    }
}

func (c *RedisStockCache) DeleteAll() error {
    keys, err := c.client.Keys(ctx, "*").Result()
    if err != nil {
        return fmt.Errorf("failed to get all keys: %w", err)
    }

    for _, key := range keys {
        if err := c.client.Del(ctx, key).Err(); err != nil {
            return fmt.Errorf("failed to delete key %s: %w", key, err)
        }
    }

    return nil
}
    