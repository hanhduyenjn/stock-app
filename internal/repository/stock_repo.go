package repository

import (
	"database/sql"
	"fmt"
	"stock-app/internal/entity"
	"time"
)

// StockRepo defines the interface for stock data operations.
type StockRepo interface {
	InsertIntradayData(symbol, timestamp, open, high, low, close, volume string) error
	InsertDailyData(symbol, date, open, high, low, close, volume string) error
	GetAllHistoricalData(startTime time.Time, endTime time.Time) (map[string][]*entity.StockQuote, error)
	GetHistoricalData(symbol string, startTime time.Time, endTime time.Time) ([]*entity.StockQuote, error)
	GetAllLatestData() (map[string]*entity.StockQuote, error)
	GetLatestIntradayDataTimestamp(symbol string) (string, error)
	GetLatestDailyDataDate(symbol string) (string, error)
	CreateTables() error
}

// StockRepoImpl provides methods for accessing and manipulating stock data in the database.
type StockRepoImpl struct {
	db *sql.DB
}

// NewStockRepo creates a new instance of StockRepoImpl.
func NewStockRepo(db *sql.DB) StockRepo {
	return &StockRepoImpl{db: db}
}

// InsertIntradayData inserts intraday stock data into the database.
func (repo *StockRepoImpl) InsertIntradayData(symbol, timestamp, open, high, low, close, volume string) error {
	query := `
        INSERT INTO stock_intraday_data (symbol, timestamp, open, high, low, close, volume)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (symbol, timestamp) DO UPDATE 
        SET open = EXCLUDED.open, 
            high = EXCLUDED.high, 
            low = EXCLUDED.low, 
            close = EXCLUDED.close, 
            volume = EXCLUDED.volume;`

	_, err := repo.db.Exec(query, symbol, timestamp, open, high, low, close, volume)
	if err != nil {
		return fmt.Errorf("error inserting intraday data for %s: %w", symbol, err)
	}
	return nil
}

// InsertDailyData inserts daily stock data into the database.
func (repo *StockRepoImpl) InsertDailyData(symbol, date, open, high, low, close, volume string) error {
	ts, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("error parsing date: %w", err)
	}

	query := `
        INSERT INTO stock_daily_data (symbol, date, open, high, low, close, volume)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (symbol, date) DO UPDATE 
        SET open = EXCLUDED.open, 
            high = EXCLUDED.high, 
            low = EXCLUDED.low, 
            close = EXCLUDED.close, 
            volume = EXCLUDED.volume;`

	_, err = repo.db.Exec(query, symbol, ts, open, high, low, close, volume)
	if err != nil {
		return fmt.Errorf("error inserting daily data for %s: %w", symbol, err)
	}
	return nil
}

func (repo *StockRepoImpl) GetAllHistoricalData(startTime time.Time, endTime time.Time) (map[string][]*entity.StockQuote, error) {
	query := `
        WITH intraday_data AS (
            SELECT 
                symbol,
                timestamp,
                open AS open_price,
                high AS high_price,
                low AS low_price,
                close AS price,
                volume,
                DATE(timestamp) AS intraday_date
            FROM stock_intraday_data
            WHERE timestamp BETWEEN $1 AND $2
        ),
        previous_day_data AS (
            SELECT 
                sdd.symbol,
                sdd.date AS prev_date,
                sdd.close AS prev_close
            FROM stock_daily_data sdd
        )

        SELECT
            sid.symbol,
            sid.price,
            (sid.price - pdd.prev_close) AS change,
            ((sid.price - pdd.prev_close) / pdd.prev_close * 100) AS change_percentage,
            sid.high_price,
            sid.low_price,
            sid.open_price,
            pdd.prev_close,
            sid.volume,
            sid.timestamp
        FROM intraday_data sid
        JOIN previous_day_data pdd
        ON sid.symbol = pdd.symbol 
        AND pdd.prev_date = sid.intraday_date - INTERVAL '1 day';

    `

	rows, err := repo.db.Query(query, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("error querying latest intraday data: %w", err)
	}
	defer rows.Close()

	// Map to store stock symbol to its list of StockQuote objects
	stockQuotesMap := make(map[string][]*entity.StockQuote)

	for rows.Next() {
		var quote entity.StockQuote
		err := rows.Scan(
			&quote.Symbol,
			&quote.Price,
			&quote.Change,
			&quote.ChangePercentage,
			&quote.HighPrice,
			&quote.LowPrice,
			&quote.OpenPrice,
			&quote.PrevClose,
			&quote.Volume,
			&quote.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Append the quote to the corresponding symbol in the map
		stockQuotesMap[quote.Symbol] = append(stockQuotesMap[quote.Symbol], &quote)
	}

	// Check for errors after the loop
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	// print length of each symbol
	for k, v := range stockQuotesMap {
		fmt.Printf("Symbol: %s, Length: %d\n", k, len(v))
	}
	return stockQuotesMap, nil
}

func (repo *StockRepoImpl) GetHistoricalData(symbol string, startTime time.Time, endTime time.Time) ([]*entity.StockQuote, error) {
    query := `
        WITH intraday_data AS (
            SELECT 
                symbol,
                timestamp,
                open AS open_price,
                high AS high_price,
                low AS low_price,
                close AS price,
                volume,
                DATE(timestamp) AS intraday_date
            FROM stock_intraday_data
            WHERE timestamp BETWEEN $1 AND $2
            AND symbol = $3
        ),
        previous_day_data AS (
            SELECT 
                sdd.symbol,
                sdd.date AS prev_date,
                sdd.close AS prev_close
            FROM stock_daily_data sdd
        )

        SELECT
            sid.symbol,
            sid.price,
            (sid.price - pdd.prev_close) AS change,
            ((sid.price - pdd.prev_close) / pdd.prev_close * 100) AS change_percentage,
            sid.high_price,
            sid.low_price,
            sid.open_price,
            pdd.prev_close,
            sid.volume,
            sid.timestamp
        FROM intraday_data sid
        JOIN previous_day_data pdd
        ON pdd.prev_date = sid.intraday_date - INTERVAL '1 day';
    `

    // Execute the query
    rows, err := repo.db.Query(query, startTime, endTime, symbol)
    if err != nil {
        return nil, fmt.Errorf("error querying historical intraday data for %s: %w", symbol, err)
    }
    defer rows.Close()

    // Prepare slice to hold results
    var stockQuotes []*entity.StockQuote

    // Iterate over rows
    for rows.Next() {
        var quote entity.StockQuote
        if err := rows.Scan(
            &quote.Symbol,
            &quote.Price,
            &quote.Change,
            &quote.ChangePercentage,
            &quote.HighPrice,
            &quote.LowPrice,
            &quote.OpenPrice,
            &quote.PrevClose,
            &quote.Volume,
            &quote.Timestamp,
        ); err != nil {
            return nil, fmt.Errorf("error scanning row for symbol %s: %w", symbol, err)
        }

        stockQuotes = append(stockQuotes, &quote)
    }

    // Check if there was an error during row iteration
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows for symbol %s: %w", symbol, err)
    }

    fmt.Printf("Fetched %d stock quotes for symbol: %s\n", len(stockQuotes), symbol)
    return stockQuotes, nil
}

func (repo *StockRepoImpl) GetAllLatestData() (map[string]*entity.StockQuote, error) {
	query := `
        WITH latest_intraday_data AS (
            SELECT 
                symbol,
                timestamp,
                open AS open_price,
                high AS high_price,
                low AS low_price,
                close AS price,
                volume
            FROM stock_intraday_data
            WHERE (symbol, timestamp) IN (
                SELECT symbol, MAX(timestamp)
                FROM stock_intraday_data
                GROUP BY symbol
            )
        ),
        previous_day_data AS (
            SELECT 
                sdd.symbol,
                sdd.close AS prev_close
            FROM stock_daily_data sdd
            JOIN (
                SELECT 
                    symbol, 
                    MAX(date) AS max_date
                FROM stock_daily_data
                WHERE date < CURRENT_DATE
                GROUP BY symbol
            ) prev_data
            ON sdd.symbol = prev_data.symbol AND sdd.date = prev_data.max_date
        )

        SELECT
            lid.symbol,
            lid.price,
            (lid.price - pdd.prev_close) AS change,
            ((lid.price - pdd.prev_close) / pdd.prev_close * 100) AS change_percentage,
            lid.high_price,
            lid.low_price,
            lid.open_price,
            pdd.prev_close,
            lid.volume,
            lid.timestamp
        FROM latest_intraday_data lid
        JOIN previous_day_data pdd
        ON lid.symbol = pdd.symbol;
`

	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying latest intraday data: %w", err)
	}
	defer rows.Close()

	latestQuotesMap := make(map[string]*entity.StockQuote)

	for rows.Next() {
		var quote entity.StockQuote
		err := rows.Scan(
			&quote.Symbol,
			&quote.Price,
			&quote.Change,
			&quote.ChangePercentage,
			&quote.HighPrice,
			&quote.LowPrice,
			&quote.OpenPrice,
			&quote.PrevClose,
			&quote.Volume,
			&quote.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		latestQuotesMap[quote.Symbol] = &quote
	}

	// Check for errors after the loop
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return latestQuotesMap, nil
}

// GetLatestIntradayDataTimestamp retrieves the latest intraday data timestamp for a given symbol.
func (repo *StockRepoImpl) GetLatestIntradayDataTimestamp(symbol string) (string, error) {
	query := `
        SELECT MAX(timestamp) 
        FROM stock_intraday_data 
        WHERE symbol = $1;`

	var timestamp sql.NullTime
	err := repo.db.QueryRow(query, symbol).Scan(&timestamp)
	if err != nil {
		return "", fmt.Errorf("error fetching latest timestamp for %s: %w", symbol, err)
	}
	if !timestamp.Valid {
		return "", nil
	}
	return timestamp.Time.Format("2006-01-02 15:04:05"), nil
}

// GetLatestDailyDataDate retrieves the latest daily data date for a given symbol.
func (repo *StockRepoImpl) GetLatestDailyDataDate(symbol string) (string, error) {
	query := `
        SELECT MAX(date) 
        FROM stock_daily_data 
        WHERE symbol = $1;`

	var date sql.NullTime
	err := repo.db.QueryRow(query, symbol).Scan(&date)
	if err != nil {
		return "", fmt.Errorf("error fetching latest date for %s: %w", symbol, err)
	}
	if !date.Valid {
		return "", nil
	}
	return date.Time.Format("2006-01-02"), nil
}

// CreateTables creates the stock_intraday_data and stock_daily_data tables if they do not exist.
func (repo *StockRepoImpl) CreateTables() error {
	intradayTableQuery := `
    CREATE TABLE IF NOT EXISTS stock_intraday_data (
        symbol VARCHAR(20) NOT NULL,
        timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        open NUMERIC(12,6),
        high NUMERIC(12,6),
        low NUMERIC(12,6),
        close NUMERIC(12,6),
        volume NUMERIC(12,2),
        PRIMARY KEY (symbol, timestamp)
    );`

	dailyTableQuery := `
    CREATE TABLE IF NOT EXISTS stock_daily_data (
        symbol VARCHAR(20) NOT NULL,
        date DATE NOT NULL,
        open NUMERIC(10,2) NOT NULL,
        high NUMERIC(10,2) NOT NULL,
        low NUMERIC(10,2) NOT NULL,
        close NUMERIC(10,2) NOT NULL,
        volume NUMERIC(12,2),
        PRIMARY KEY (symbol, date)
    );`

	// Execute the intraday table creation query
	_, err := repo.db.Exec(intradayTableQuery)
	if err != nil {
		return fmt.Errorf("error creating stock_intraday_data table: %w", err)
	}

	// Execute the daily table creation query
	_, err = repo.db.Exec(dailyTableQuery)
	if err != nil {
		return fmt.Errorf("error creating stock_daily_data table: %w", err)
	}

	return nil
}
