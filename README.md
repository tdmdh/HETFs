# halalctl

> **A fast, open-source CLI for discovering and analyzing Halal-compliant ETFs using the Interactive Brokers Client Portal API.**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![IBKR API](https://img.shields.io/badge/IBKR-Client%20Portal%20API-blue)](https://interactivebrokers.github.io/cpwebapi/)

---

## Overview

`halalctl` is a command-line tool that connects directly to the **Interactive Brokers Client Portal Gateway REST API** to:

- 🔍 **Search** for ETF instruments by ticker symbol
- 📊 **Scan** markets for top-performing ETFs in real time
- 💾 **Cache** contract metadata locally (SQLite) for fast repeat lookups
- ⚡ **Rate-limit safely** — enforces IBKR's strict 10 req/s limit with automatic exponential backoff

Built for developers and traders who want a lightweight, scriptable interface to IBKR market data.

---

## Features

| Feature | Details |
|---|---|
| ETF Search | Maps ticker symbols to IBKR `conid` values via `/secdef/search` |
| Market Scanner | Runs configurable scans (e.g. `TOP_PERC_GAIN`) via `/scanner/run` |
| Real-time Prices | Fetches live snapshots via `/marketdata/snapshot` with 100-conid chunking |
| Session Keepalive | Background goroutine tickles the gateway every 45s to prevent session expiry |
| Local Cache | SQLite stores `conid` mappings so repeat lookups are instant and offline |
| Rate Limiting | `golang.org/x/time/rate` enforces ≤9 req/s with burst=1 |
| Retry Logic | Exponential backoff (1s → 2s → 4s) on HTTP 429 responses |

---

## Requirements

- [Go 1.22+](https://go.dev/dl/)
- An **Interactive Brokers** account (paper or live)
- **IBKR Client Portal Gateway** running locally ([download here](https://interactivebrokers.github.io/cpwebapi/))
- Java 8+ (for the Gateway)

---

## Installation

```bash
git clone https://github.com/tdmdh/HETFs.git
cd HETFs
go build -o bin/halalctl .
```

Or run directly without building:

```bash
go run . <command>
```

---

## Gateway Setup

`halalctl` communicates with the **IBKR Client Portal Gateway**, which must be running locally.

### 1. Start the Gateway

```powershell
# Windows (from the extracted clientportal.gw folder)
bin\run.bat root\conf.yaml
```

```bash
# macOS / Linux
bin/run.sh root/conf.yaml
```

### 2. Log in via Browser

Open your browser and navigate to:

```
https://127.0.0.1:5000
```

> ⚠️ Use `127.0.0.1` explicitly — **not** `localhost`. The Gateway ties sessions to specific IP addresses, and IPv4/IPv6 resolution differences will cause authentication to fail.

Accept the self-signed certificate warning and log in with your IBKR credentials. You should see **"Client login succeeds"**.

### 3. Extract the Session Cookie

`halalctl` needs to share your browser's authenticated session. After logging in:

1. Open **DevTools** (`F12`) → **Application** tab → **Cookies** → `https://127.0.0.1`
2. Copy the **Value** of the `x-sess-uuid` cookie
3. Create a `.env` file in the project root:

```env
SESSION_COOKIE=x-sess-uuid=<paste-value-here>
```

`.env` is gitignored — your credentials are never committed.

---

## Usage

```bash
# Verify gateway connectivity and authentication
go run . init

# Search for an ETF by ticker symbol
go run . search VWRL

# Run a market scanner
go run . scan TOP_PERC_GAIN --location ETF.US.MAJOR
```

### `init` — Verify Gateway

```
$ go run . init
⏳ Polling Gateway at https://127.0.0.1:5000...
✅ Gateway is AUTHENTICATED. Ready to serve requests.
```

If the gateway is not running, `init` will attempt to start it via `docker-compose` using credentials from `.env`.

### `search` — Find an ETF

```
$ go run . search VWRL
Searching for 'VWRL'...
Found Match: VANGUARD FTSE ALL-WORLD U (conid: 114421160, Exchange: SMART)
Saved VWRL to local cache.
```

### `scan` — Market Scanner

```
$ go run . scan TOP_PERC_GAIN --location ETF.US.MAJOR
Running ETF market scanner 'TOP_PERC_GAIN' at 'ETF.US.MAJOR'...
Found 13 ETFs. Fetching real-time market data...

 Symbol |          Company Name          | Last Price | % Change
--------|--------------------------------|------------|----------
 NVDL   | GRANITESHARES 2X LONG NVDA DAI | $46.88     | 9.87%
 DPST   | DIREXION DAILY REGIONAL B 3X B | $104.29    | 8.13%
 KRE    | SPDR S&P REGIONAL BANKING ETF  | $52.09     | 3.12%
```

---

## Configuration

All options are set via flags or environment variables.

| Flag | Env Var | Default | Description |
|---|---|---|---|
| `--gateway-url` | — | `https://127.0.0.1:5000` | URL of the IBKR Client Portal Gateway |
| `--db` | — | `halalctl.db` | Path to local SQLite cache |
| — | `SESSION_COOKIE` | — | Browser session cookie (from `.env`) |
| — | `TWS_USERID` | — | IBKR username (for Docker fallback) |
| — | `TWS_PASSWORD` | — | IBKR password (for Docker fallback) |
| — | `TRADING_MODE` | `paper` | `live` or `paper` (for Docker fallback) |

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

---

## Architecture

```
halalctl/
├── cmd/
│   ├── root.go        # Cobra root: auth check + tickle goroutine
│   ├── init.go        # Gateway connectivity + session setup guide
│   ├── search.go      # ETF symbol → conid lookup
│   └── scan.go        # Market scanner + table output
├── internal/
│   ├── ibkr/
│   │   ├── client.go     # HTTP client: rate limiter + cookie injection + retry
│   │   ├── auth.go       # CheckAuthStatus + Tickle
│   │   ├── secdef.go     # /secdef/search wrapper
│   │   ├── scanner.go    # /scanner/run wrapper
│   │   └── marketdata.go # /marketdata/snapshot with 100-conid chunking
│   ├── domain/
│   │   └── etf.go        # Contract domain model
│   └── storage/
│       ├── sqlite.go     # SQLite connection + schema
│       └── etf_repo.go   # Contract cache repository
├── main.go
├── docker-compose.yml    # Optional: gnzsnz/ib-gateway fallback
├── .env.example
└── Makefile
```

---

## Docker Fallback

If the Gateway is not running, `halalctl init` will attempt to start it automatically using Docker.

Configure your `.env`:

```env
TWS_USERID=your_ibkr_username
TWS_PASSWORD=your_ibkr_password
TRADING_MODE=paper
```

Then run:

```bash
docker-compose up -d
```

> **Note:** The Docker image (`gnzsnz/ib-gateway`) uses IBC to auto-login and runs the classic TCP Gateway. It does **not** expose the Client Portal REST API on port 5000. For full CLI functionality, use the official Client Portal Gateway from IBKR directly.

---

## Development

```bash
# Run all tests
go test ./...

# Build binary
go build -o bin/halalctl .

# Format code
go fmt ./...

# Tidy dependencies
go mod tidy
```

---

## Contributing

Contributions are welcome! Please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'feat: add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

---

## Disclaimer

This project is an **unofficial, community-maintained** tool and is **not affiliated with or endorsed by Interactive Brokers**. Use at your own risk. Always review IBKR's [API Terms of Service](https://www.interactivebrokers.com/en/trading/terms.php) before using this tool in a production or live-trading environment.

---

## License

[MIT](./LICENSE)
