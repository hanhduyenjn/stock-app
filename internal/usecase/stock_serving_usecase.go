package usecase

import (
	"fmt"
	"time"

	"stock-app/internal/cache"
	"stock-app/internal/entity"
	"stock-app/internal/repository"
	"stock-app/pkg/config"
)

// StockServingUseCase defines the business logic related to stock data.
type StockServingUseCase struct {
	stockRepo       repository.StockRepo
	stockCache      cache.StockCache
	latestQuoteData *entity.LatestQuoteData
}

// NewStockServingUseCase creates a new instance of StockServingUseCase.
func NewStockServingUseCase(
	stockRepo repository.StockRepo,
	stockCache cache.StockCache,
	latestQuoteData *entity.LatestQuoteData,
) *StockServingUseCase {
	return &StockServingUseCase{
		stockRepo:       stockRepo,
		stockCache:      stockCache,
		latestQuoteData: latestQuoteData,
	}
}

// GetLatestQuote retrieves the stock quote by symbol.
func (uc *StockServingUseCase) GetQuote(symbol string, start, end time.Time) ([]*entity.StockQuote, error) {
	// Check cache for quotes within the specified time range
	quotes, found := uc.stockCache.Get(symbol, start, end)
	if !found || len(quotes) == 0 {
		// get from stockRepo
		quotes, err := uc.stockRepo.GetHistoricalData(symbol, start, end)
		if err != nil {
			return nil, fmt.Errorf("failed to get historical data by symbol and range: %w", err)
		}
		if len(quotes) > 0 {
			if err := uc.stockCache.Set(symbol, quotes, config.AppConfig.CacheShortTTL); err != nil {
				return nil, fmt.Errorf("failed to set historical data in cache: %w", err)
			}
		}
	}
	return quotes, nil
}

// GetAllQuotes retrieves stock data for all symbols.
func (uc *StockServingUseCase) GetAllQuotes() (map[string]*entity.StockQuote, error) {
	// Check cache for latest quotes of all symbols
	quotes, found := uc.stockCache.GetAllLatest()
	if !found {
		// get from stockRepo
		quotes, err := uc.stockRepo.GetAllLatestData()
		if err != nil {
			return nil, fmt.Errorf("failed to get all latest data: %w", err)
		}
		if err := uc.stockCache.SetAllLatest(quotes, config.AppConfig.CacheShortTTL); err != nil {
			return nil, fmt.Errorf("failed to set all latest data in cache: %w", err)
		}

	}
	return quotes, nil
}

// func (uc *StockServingUseCase) GetTrades(symbol, timeRange string) ([]*entity.Trade, error) {
//     // if symbol == "" || timeRange == "" {
//     //     return nil, fmt.Errorf("symbol and time range are required")
//     // }

//     // fmt.Printf("Fetching trades for symbol: %s, time range: %s\n", symbol, timeRange)

//     // trades, err := uc.stockCache.GetTrades(symbol, timeRange)
//     // if err != nil {
//     //     return nil, fmt.Errorf("failed to get trades from cache: %w", err)
//     // }

//     // fmt.Printf("Trades for symbol %s, time range %s: %+v\n", symbol, timeRange, trades)
//     // return trades, nil
// }

// func (uc *StockServingUseCase) GetCompanyProfile(symbol string) (*entity.CompanyProfile, error) {
//     // if symbol == "" {
//     //     return nil, fmt.Errorf("symbol is required")
//     // }

//     // fmt.Printf("Fetching company profile for symbol: %s\n", symbol)

//     // profile, err := uc.stockCache.GetCompanyProfile(symbol)
//     // if err != nil {
//     //     return nil, fmt.Errorf("failed to get company profile from cache: %w", err)
//     // }

//     // fmt.Printf("Company profile for symbol %s: %+v\n", symbol, profile)
//     // return profile, nil
// }

// func (uc *StockServingUseCase) GetFinancials(symbol string) (*entity.Financials, error) {
//     // if symbol == "" {
//     //     return nil, fmt.Errorf("symbol is required")
//     // }

//     // fmt.Printf("Fetching financials for symbol: %s\n", symbol)

//     // financials, err := uc.stockCache.GetFinancials(symbol)
//     // if err != nil {
//     //     return nil, fmt.Errorf("failed to get financials from cache: %w", err)
//     // }

//     // fmt.Printf("Financials for symbol %s: %+v\n", symbol, financials)
//     // return financials, nil
// }
