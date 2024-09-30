package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"stock-app/internal/usecase"
)

// StockHandler defines the business logic related to stock data.
type StockHandler struct {
	stockUseCase *usecase.StockServingUseCase
}

// NewStockHandler creates a new instance of StockHandler.
func NewStockHandler(stockUseCase *usecase.StockServingUseCase) *StockHandler {
	return &StockHandler{
		stockUseCase: stockUseCase,
	}
}

// GetAllQuotes handles GET requests to retrieve all stock data.
func (sh *StockHandler) GetAllQuotes(c *gin.Context) {
	stockList, err := sh.stockUseCase.GetAllQuotes() 
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get list of stocks: %v", err)})
		return
	}
	c.JSON(http.StatusOK, stockList)
}

// Request model for getting stock by symbol
type GetQuoteRequest struct {
	Symbol string `uri:"symbol" binding:"required,alpha"`
}

// GetQuote handles GET requests to retrieve stock data by symbol.
func (sh *StockHandler) GetQuote(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is a required query parameter"})
        return
    }

	startTimeStr := c.Query("start")
	endTimeStr := c.Query("end")

	var startTime, endTime time.Time
	var err error

	if startTimeStr == "" {
		startTime = time.Now().AddDate(0, 0, -1)
	} else {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start time format"})
			return
		}
	}

	if endTimeStr == "" {
		endTime = time.Now()
	} else {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end time format"})
			return
		}
	}

	stock, err := sh.stockUseCase.GetQuote(symbol, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get stock data by symbol: %v", err)})
		return
	}
	if stock == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("stock not found for symbol: %s", symbol)})
		return
	}
	c.JSON(http.StatusOK, stock)
}

// func (h *StockHandler) GetTrades(c *gin.Context) {
//     symbol := c.Query("symbol")
//     timeRange := c.Query("range")
//     if symbol == "" || timeRange == "" {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "symbol and range are required query parameters"})
//         return
//     }
//     trades, err := h.stockUseCase.GetTrades(symbol, timeRange)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//         return
//     }
//     c.JSON(http.StatusOK, trades)
// }

// func (h *StockHandler) GetCompanyProfile(c *gin.Context) {
//     symbol := c.Query("symbol")
//     if symbol == "" {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is a required query parameter"})
//         return
//     }
//     profile, err := h.stockUseCase.GetCompanyProfile(symbol)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//         return
//     }
//     c.JSON(http.StatusOK, profile)
// }

// func (h *StockHandler) GetFinancials(c *gin.Context) {
//     symbol := c.Query("symbol")
//     if symbol == "" {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is a required query parameter"})
//         return
//     }
//     financials, err := h.stockUseCase.GetFinancials(symbol)
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//         return
//     }
//     c.JSON(http.StatusOK, financials)
// }



