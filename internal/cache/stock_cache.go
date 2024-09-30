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
    Get(symbol string, startTime, endTime time.Time) ([]*entity.StockQuote, bool)
    GetAll(startTime, endTime time.Time) (map[string][]*entity.StockQuote, bool)
    GetAllLatest() (map[string]*entity.StockQuote, bool)
    Set(symbol string, stock []*entity.StockQuote, expiration time.Duration) error
    SetAll(stocks map[string][]*entity.StockQuote, expiration time.Duration) error
    SetLatest(symbol string, stock *entity.StockQuote, expiration time.Duration)
    SetAllLatest(stocks map[string]*entity.StockQuote, expiration time.Duration) error
    DeleteAll() error
}

// RedisStockCache is a Redis-backed cache for stock data.
type RedisStockCache struct {
    client *redis.Client
}

// NewStockCache creates a new RedisStockCache instance.
func NewStockCache(redisAddr string) StockCache {
    rdb := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    return &RedisStockCache{client: rdb}
}

// Get retrieves stock data from the cache by symbol for a given time range.
func (c *RedisStockCache) Get(symbol string, startTime, endTime time.Time) ([]*entity.StockQuote, bool) {
    key := fmt.Sprintf("stock:%s:history", symbol)
    stockData, err := c.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
        Min: fmt.Sprintf("%d", startTime.Unix()),
        Max: fmt.Sprintf("%d", endTime.Unix()),
    }).Result()

    if err != nil || len(stockData) == 0 {
        return nil, false // Cache miss or Redis error
    }

    return c.unmarshalStockQuotes(stockData), true
}

// GetAll retrieves all stocks from the cache.
func (c *RedisStockCache) GetAll(startTime, endTime time.Time) (map[string][]*entity.StockQuote, bool) {
    stocks := make(map[string][]*entity.StockQuote)
    keys, err := c.client.Keys(ctx, "stock:*:history").Result()
    if err != nil {
        return nil, false // Redis error
    }

    for _, key := range keys {
        symbol := key[6 : len(key)-8] // Extract the symbol from the key
        if stockQuotes, found := c.Get(symbol, startTime, endTime); found {
            stocks[symbol] = stockQuotes
        }
    }

    return stocks, len(stocks) > 0
}

// GetAllLatest retrieves the latest stock data from the cache.
func (c *RedisStockCache) GetAllLatest() (map[string]*entity.StockQuote, bool) {
    stocks := make(map[string]*entity.StockQuote)
    keys, err := c.client.Keys(ctx, "stock:*:history").Result()
    if err != nil {
        return nil, false // Redis error
    }

    for _, key := range keys {
        if stockData, err := c.client.ZRevRange(ctx, key, 0, 0).Result(); err == nil && len(stockData) > 0 {
            var stock entity.StockQuote
            if err := json.Unmarshal([]byte(stockData[0]), &stock); err == nil {
                symbol := key[6 : len(key)-8] // Extract the symbol from the key
                stocks[symbol] = &stock
            } else {
                fmt.Printf("Failed to unmarshal stock data: %v\n", err)
            }
        }
    }

    return stocks, len(stocks) > 0
}

// Set stores stock data in the cache with an optional expiration time.
func (c *RedisStockCache) Set(symbol string, stock []*entity.StockQuote, expiration time.Duration) error {
    key := fmt.Sprintf("stock:%s:history", symbol)
    
    // Prepare the []*redis.Z data
    zData := c.prepareZData(stock) 

    if err := c.client.ZAdd(ctx, key, zData...).Err(); err != nil {
        fmt.Printf("Failed to cache stock %s: %v\n", symbol, err)
        return err
    }

    // Set expiration for the sorted set if specified
    if expiration > 0 {
        c.client.Expire(ctx, key, expiration)
    }
    
    fmt.Printf("Successfully cached all stock data for %s\n", symbol)
    return nil
}


// SetAll stores multiple stocks in the cache with an optional expiration time.
func (c *RedisStockCache) SetAll(stocks map[string][]*entity.StockQuote, expiration time.Duration) error {
    var wg sync.WaitGroup
    for symbol, stockValues := range stocks {
        wg.Add(1)
        go func(symbol string, stockValues []*entity.StockQuote) {
            defer wg.Done()
            _ = c.Set(symbol, stockValues, expiration) // Ignore errors for simplicity
        }(symbol, stockValues)
    }
    wg.Wait()
    return nil
}

// SetLatest stores a single stock in the cache.
func (c *RedisStockCache) SetLatest(symbol string, stock *entity.StockQuote, expiration time.Duration) {
    key := fmt.Sprintf("stock:%s:history", symbol)
    stockJSON, err := json.Marshal(stock)
    if err != nil {
        fmt.Printf("Failed to marshal stock data for %s: %v\n", symbol, err)
        return
    }

    if err := c.client.ZAdd(ctx, key, &redis.Z{
        Score:  float64(stock.Timestamp.Unix()),
        Member: stockJSON,
    }).Err(); err != nil {
        fmt.Printf("Failed to cache stock %s: %v\n", symbol, err)
    } else {
        fmt.Printf("Successfully cached stock %s\n", symbol)
    }

    // Set expiration if specified
    if expiration > 0 {
        c.client.Expire(ctx, key, expiration)
    }
}

// SetAllLatest stores multiple stocks in the cache using sorted sets.
func (c *RedisStockCache) SetAllLatest(stocks map[string]*entity.StockQuote, expiration time.Duration) error {
    var wg sync.WaitGroup
    for symbol, stock := range stocks {
        wg.Add(1)
        go func(symbol string, stock *entity.StockQuote) {
            defer wg.Done()
            c.SetLatest(symbol, stock, expiration)
        }(symbol, stock)
    }
    wg.Wait()
    return nil
}

// DeleteAll deletes all stock data from the cache.
func (c *RedisStockCache) DeleteAll() error {
    keys, err := c.client.Keys(ctx, "stock:*:history").Result()
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

// Helper function to unmarshal stock quotes from JSON data.
func (c *RedisStockCache) unmarshalStockQuotes(stockData []string) []*entity.StockQuote {
    var stockQuotes []*entity.StockQuote
    for _, stockJSON := range stockData {
        var stock entity.StockQuote
        if err := json.Unmarshal([]byte(stockJSON), &stock); err != nil {
            fmt.Printf("Failed to unmarshal stock data: %v\n", err)
            continue // Skip on unmarshalling error
        }
        stockQuotes = append(stockQuotes, &stock)
    }
    return stockQuotes
}

// Helper function to prepare Redis Z data for batch insertion.
func (c *RedisStockCache) prepareZData(stock []*entity.StockQuote) []*redis.Z {
    var zData []*redis.Z
    for _, s := range stock {
        stockJSON, err := json.Marshal(s)
        if err != nil {
            fmt.Printf("Failed to marshal stock data: %v\n", err)
            continue
        }
        zData = append(zData, &redis.Z{
            Score:  float64(s.Timestamp.Unix()), // Use timestamp as score
            Member: stockJSON,                   // Store the marshaled JSON as the member
        })
    }
    return zData
}

