package entity

import (
    "sync"
    "time"
)

// AlphaVantage


type TimeSeriesData struct {
    Open   string `json:"1. open" validate:"required"`
    High   string `json:"2. high" validate:"required"`
    Low    string `json:"3. low" validate:"required"`
    Close  string `json:"4. close" validate:"required"`
    Volume string `json:"5. volume" validate:"required"`
}

type MetaDataIntraday struct {
    Information   string `json:"1. Information" validate:"required"`
    Symbol        string `json:"2. Symbol" validate:"required"`
    LastRefreshed string `json:"3. Last Refreshed" validate:"required"`
    Interval      string `json:"4. Interval" validate:"required"`
    OutputSize    string `json:"5. Output Size" validate:"required"`
    TimeZone      string `json:"6. Time Zone" validate:"required"`
}


type TSIntradayResponse struct {
    MetaData   MetaDataIntraday          `json:"Meta Data" validate:"required,dive"`
    TimeSeries map[string]TimeSeriesData `json:"Time Series (1min)" validate:"required,dive"`
}

type MetaDataDaily struct {
    Information   string `json:"1. Information" validate:"required"`
    Symbol        string `json:"2. Symbol" validate:"required"`
    LastRefreshed string `json:"3. Last Refreshed" validate:"required"`
    OutputSize    string `json:"4. Output Size" validate:"required"`
    TimeZone      string `json:"5. Time Zone" validate:"required"`
}

type TSDailyResponse struct {
    MetaData   MetaDataDaily             `json:"Meta Data" validate:"required,dive"`
    TimeSeries map[string]TimeSeriesData `json:"Time Series (Daily)" validate:"required,dive"`
}

type MetaData struct {
    Information   string `json:"1. Information" validate:"required"`
    Symbol        string `json:"2. Symbol" validate:"required"`
    LastRefreshed string `json:"3. Last Refreshed" validate:"required"`
    Interval      string `json:"4. Interval" validate:"required"`
    OutputSize    string `json:"5. Output Size" validate:"required"`
    TimeZone      string `json:"6. Time Zone" validate:"required"`
}

// finnhub
type StockQuote struct {
    Symbol           string  `json:"s"`
    Price            float64 `json:"c"`
    Change           float64 `json:"d"`
    ChangePercentage float64 `json:"dp"`
    HighPrice        float64 `json:"h"`
    LowPrice         float64 `json:"l"`
    OpenPrice        float64 `json:"o"`
    PrevClose        float64 `json:"pc"`
    Volume           float64  `json:"v"`
    Timestamp        time.Time  `json:"t"`
}

// LatestQuoteData holds real-time stock data in memory.
type LatestQuoteData struct {
    StockData map[string]*StockQuote `json:"StockData"`
    Mu        sync.RWMutex           `json:"Mu"`
}