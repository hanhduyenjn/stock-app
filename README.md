
# Stock App

This project is a stock application that fetches and processes stock data from various APIs and serves it through a Go-based server.


## Setting Up the Project

### Prerequisites
- **Go:** Ensure you have Go installed. You can download it from the [official Go website](https://golang.org/doc/install).
- **Docker:** Install Docker by following the instructions [here](https://docs.docker.com/get-docker/).
- **Docker Compose:** Install Docker Compose by following the instructions [here](https://docs.docker.com/compose/install/).

### Creating Database with Docker
To create a PostgreSQL database using Docker, you can run the following command in your terminal:

```sh
docker run --name stock-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=stockdatabase -p 5432:5432 -d postgres:13
```

To run Redis using Docker, you can execute:

```sh
docker run --name stock-redis -p 6379:6379 -d redis:alpine
```

## Environment Variables

Create a `.env` file in the root directory with the following content:

```env
SYMBOL_LIST=TSLA,GOOGL,AMZN,MSFT#,META,NVDA,BABA,AMD,INTC,CRM,NFLX,TWTR,BA,WMT,DIS,PFE,XOM,JPM,V,MA,CSCO,T,KO,HD,NKE,CVX,MCD,UNH,WFC,ABT,MDT,LLY,ORCL,BMY,C,GS,AIG,UPS,F,TMO,CVS,ABBV,AMGN,SPY,TSM,NIO,GILD,HCA,SQ,RBLX,SHOP,U,PLTR,PINS,Roku,BYND,FUBO,NCLH,AAL,CCL,DAL,UAL,LUV,MGM,CROX,LULU,HIMS,L,GME,AMC,PLTR,TSLA,RBLX,NIO,SNAP,Z,GOOG,NVDA,SHOP,PDD,BABA,ADBE,INTC,QCOM,XOM,CVX,MCD,MS,AXP,AAPL,TWLO,SHOP,RBLX,PLTR,HYLN,QS,BLNK

# Alphavantage
ALPHA_VANTAGE_API_KEY=#Get free API key here: https://www.alphavantage.co/support/#api-key
TIMESERIES_ENDPOINT=https://www.alphavantage.co/query?outputsize=full&extended_hours=false

# Finnhub
FINHUBB_API_KEY=#Get free API key here: https://finnhub.io/dashboard
REAL_TIME_TRADES_ENDPOINT=wss://ws.finnhub.io
QUOTE_ENDPOINT=https://finnhub.io/api/v1/quote
COMPANY_PROFILE_ENDPOINT=https://finnhub.io/api/v1/stock/profile2

# Database configuration
DB_USERNAME=postgres
DB_PASSWORD=mysecretpassword
DB_HOST=localhost
DB_PORT=5432
DB_NAME=stockdatabase

# Redis configuration
REDIS_HOST=localhost
REDIS_PORT=6379
CACHE_SHORT_TTL=30
CACHE_LONG_TTL=235800

# Logging settings
LOG_LEVEL=debug

# Server configuration
SERVER_PORT=8080
```

## Makefile Commands
- `make create`: Create tables in the database `stockdatabase`.
- `make refresh`: Get the latest data from API to fetch in the database.
- `make build`: Build the Go application.
- `make run`: Run the Go application.
- `make cleanup`: Clean up cache.

## Running the Application

1. **Start web server:**

   ```sh
   make all
   ```

   or 
   ```sh
   go run main.go
   ```