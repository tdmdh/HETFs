Below is a **professional open-source repository structure and README** suitable for your **IBKR ETF CLI project**. The goal is to make the project look **clean, maintainable, and production-ready**, which is important if you want contributors later.

---

# Recommended Project Structure

```
ibkr-etf-cli/
│
├── cmd/
│   ├── root.go
│   ├── init.go
│   ├── search.go
│   ├── list.go
│   └── version.go
│
├── internal/
│   ├── gateway/
│   │   ├── client.go
│   │   ├── auth.go
│   │   └── tickle.go
│   │
│   ├── ibkr/
│   │   ├── search.go
│   │   ├── scanner.go
│   │   ├── marketdata.go
│   │   └── contracts.go
│   │
│   ├── cache/
│   │   ├── cache.go
│   │   └── sqlite.go
│   │
│   ├── models/
│   │   ├── contract.go
│   │   ├── etf.go
│   │   └── marketdata.go
│   │
│   ├── rate/
│   │   └── limiter.go
│   │
│   └── output/
│       └── table.go
│
├── pkg/
│   └── utils/
│       └── logger.go
│
├── configs/
│   └── config.yaml
│
├── scripts/
│   └── start-gateway.sh
│
├── docs/
│   ├── architecture.md
│   └── ibkr-api-notes.md
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── LICENSE
└── .gitignore
```

---

# Explanation of Key Folders

## `cmd/`

Contains CLI commands created with **Cobra**.

Example commands:

```
ibkr-etf init
ibkr-etf search VOO
ibkr-etf list top-gainers
ibkr-etf list top-volume
```

---

## `internal/`

Core application logic.

### `gateway/`

Handles communication with the **IBKR Client Portal Gateway**.

Responsibilities:

* Checking authentication
* Session keep-alive (`tickle`)
* HTTP client setup

---

### `ibkr/`

Handles **actual API endpoints**.

Example modules:

```
search.go      -> secdef/search
scanner.go     -> scanner/run
marketdata.go  -> marketdata/snapshot
contracts.go   -> contract lookup
```

---

### `cache/`

Local caching system for **symbol → conid mapping**.

Options:

* JSON
* SQLite

SQLite is recommended because it scales better.

---

### `models/`

Go structs representing API responses.

Example:

```
type Contract struct {
    Conid       int    `json:"conid"`
    Symbol      string `json:"symbol"`
    CompanyName string `json:"companyName"`
}
```

---

### `rate/`

Implements **rate limiting** using

```
golang.org/x/time/rate
```

This prevents hitting the IBKR **10 requests/sec limit**.

---

### `output/`

Responsible for rendering CLI tables.

Uses:

```
github.com/olekukonko/tablewriter
```

Example output:

```
Ticker   Name                     Price   Change
------------------------------------------------
VOO      Vanguard S&P 500 ETF    485.23   +0.43%
VTI      Vanguard Total Market   263.11   +0.31%
QQQ      Invesco Nasdaq 100      512.02   +0.62%
```

---

# Example CLI Commands

### Initialize Gateway Connection

```
ibkr-etf init
```

Checks:

* Gateway running
* Authentication status
* Session validity

---

### Search ETF

```
ibkr-etf search VOO
```

Output:

```
Symbol   Exchange   Conid
-------------------------
VOO      ARCA       12345678
```

---

### List ETFs (Scanner)

```
ibkr-etf list top-gainers
```

Possible filters:

```
top-gainers
top-losers
most-active
highest-volume
```

---

# Example Makefile

```
build:
	go build -o ibkr-etf

run:
	go run main.go

lint:
	golangci-lint run

test:
	go test ./...

format:
	go fmt ./...
```

---

# Example README.md

````markdown
# IBKR ETF CLI

A fast, open-source CLI for discovering and analyzing ETFs using the Interactive Brokers Client Portal API.

## Features

- ETF symbol search
- ETF market scanners
- Real-time market data
- Local contract caching
- Rate-limit safe API usage

## Requirements

- Interactive Brokers account
- Client Portal Gateway running locally

Default gateway:

http://localhost:5000

## Installation

```bash
git clone https://github.com/yourname/ibkr-etf-cli
cd ibkr-etf-cli
go build
````

## Usage

Initialize gateway connection:

```
ibkr-etf init
```

Search ETF:

```
ibkr-etf search VOO
```

List ETFs:

```
ibkr-etf list top-gainers
```

## Architecture

See `/docs/architecture.md`.

## License

MIT

```

---

# Suggested Project Names

Since you're making this **open source**, good names could be:

**Professional sounding**

- `ETFScout`
- `IBKR-ETF-CLI`
- `ETFScanner`
- `ETFAlpha`

**Developer style**

- `etfctl`
- `etfscan`
- `ibkr-scan`
- `alpha-cli`

My favorite for this project:

**`etfctl`**

Because CLI tools often use `ctl` like:

- `kubectl`
- `systemctl`

---

# Optional Advanced Features (Later)

These would make your project **very impressive on GitHub**.

### Shariah ETF Filter

Since you mentioned halal ETFs earlier.

Add:

```

ibkr-etf list halal

```

Which checks:

- Alcohol
- Gambling
- Interest-based finance

---

### Terminal Charts

Use libraries like:

```

github.com/guptarohit/asciigraph

```

Example:

```

VOO Price (30d)

520 ┤
510 ┤
500 ┤      ╭─╮
490 ┤  ╭───╯ ╰───╮
480 ┼──╯         ╰──

```

---

# Realistic Development Time

If you work **consistently**, Joyboy:

| Phase | Time |
|------|------|
Gateway + CLI | 1–2 days |
Symbol search | 1 day |
Scanner | 2 days |
Market data | 1 day |
Caching | 1 day |
Output formatting | 1 day |

Total:

**~1 week for a solid MVP**

---

If you want, I can also show you the **exact first steps to start coding this tomorrow** (the **first 10 commits you should make** so the project grows cleanly).
```
