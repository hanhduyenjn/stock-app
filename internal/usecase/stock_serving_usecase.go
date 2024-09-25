package usecase

import (
	"context"
	"fmt"

	"stock-app/internal/cache"
	"stock-app/internal/entity"
	"stock-app/internal/repository"
)

// StockServingUseCase defines the business logic related to stock data.
type StockServingUseCase struct {
	stockRepo       repository.StockRepo
	stockCache      cache.StockCache
	latestQuoteData *entity.LatestQuoteData
}

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

// GetQuote retrieves stock data by symbol.
func (uc *StockServingUseCase) GetQuote(ctx context.Context, symbol string) (*entity.StockQuote, error) {
    // Check cache first
    quote, found := uc.stockCache.Get(symbol)
    if !found {
        return quote, fmt.Errorf("stock not found for symbol: %s", symbol)
    }
    return quote, nil
}

// GetAllQuotes retrieves stock data for all symbols.
func (uc *StockServingUseCase) GetAllQuotes(ctx context.Context) ([]*entity.StockQuote, error) {
    quotes, found := uc.stockCache.GetAll()
    if !found {
        return quotes, fmt.Errorf("stock data not found")
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
