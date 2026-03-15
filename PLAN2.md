
---

# IBKR CLI Project Roadmap

## Phase 1: Go Project Bootstrapping and Gateway Setup

The first phase focuses on establishing the project structure and the persistent authentication required by the **Client Portal Gateway**.

### CLI Framework

Use **Cobra** (`github.com/spf13/cobra`) to manage the command hierarchy.

Initialize the project with:

```
cobra-cli init
```

This creates the standard `cmd/` directory structure for CLI commands.

### API Client Choice

**Option 1 — REST API (Recommended for CLIs)**
Use:

```
github.com/gomisha/ibkr-api
```

or build a custom wrapper using Go's `net/http` for finer control.

**Option 2 — TWS API (Socket-based)**

```
github.com/scmhub/ibapi
```

### Gateway Orchestration

The CLI must detect whether the **Java Client Portal Gateway** is running on the default port:

```
5000
```

Implement an `init` command that verifies the endpoint:

```
/v1/api/iserver/auth/status
```

### Session Management

Create a background **Go routine** that keeps the session alive.

Call:

```
/v1/api/tickle
```

every **30–60 seconds** to prevent session timeouts.

---

# Phase 2: Instrument Discovery and Contract Mapping

This phase implements the logic that translates human-readable **ETF symbols** into IBKR's required **Contract Identifiers (`conid`)**.

### Symbol Search Module

Implement a `search` command using:

```
POST /v1/api/iserver/secdef/search
```

Define Go structs that match the IBKR JSON response.

Example fields:

* `conid`
* `symbol`
* `companyName`

### Filtering Logic

Although ETFs trade like stocks (`secType: "STK"`), the search results may contain duplicates.

The Go logic should prioritize entries where the exchange matches major venues such as:

* `ARCA`
* `SMART`

### Local Cache

Store mappings locally to reduce API calls.

Recommended options:

* JSON file
* SQLite (`modernc.org/sqlite`)

Purpose:

* Reduce API latency
* Avoid rate limits

---

# Phase 3: Market Scanner and ETF Listing

This phase enables **ETF discovery based on market performance**.

### Dynamic Parameters

Retrieve scanner capabilities:

```
GET /v1/api/iserver/scanner/params
```

Use Go's `encoding/xml` to parse the large XML response.

Important parameters:

* `locationCode` (example: `ETF.US.MAJOR`)
* `scanCode`

### Scanner Execution

Implement a `list` command that sends a `scan_body` to:

```
POST /v1/api/iserver/scanner/run
```

### Stock Type Filtering

Ensure the scanner request explicitly filters ETFs.

Example:

```
"instrument": "ETF"
```

or

```
stockTypeFilter = "ETF"
```

This excludes normal corporate stocks.

---

# Phase 4: Data Enrichment and Output Rendering

The final phase retrieves pricing data and displays results in a professional CLI format.

### Batch Snapshot Retrieval

Use the endpoint:

```
/v1/api/iserver/marketdata/snapshot
```

This retrieves real-time market data.

Important limitation:

* Up to **100 conids per request**

This batching is critical to stay within the:

```
10 requests per second
```

limit.

### Concurrent Enrichment

For endpoints that cannot be batched (for example):

```
/v1/api/iserver/secdef/info
```

Use Go's **worker pool pattern** combined with **rate-limited channels**.

Purpose:

* Prevent `429 Too Many Requests` errors.

### Output Formatting

Render CLI output using a table library such as:

```
github.com/olekukonko/tablewriter
```

Example columns:

* Ticker
* Name
* Last Price
* % Change

---

# Operational Governance for Go Implementations

| Requirement     | Value      | Implementation Note                                           |
| --------------- | ---------- | ------------------------------------------------------------- |
| Rate Limiting   | 10 req/sec | Use `golang.org/x/time/rate` to enforce client-side pacing    |
| Auth Status     | Boolean    | Always verify `authenticated == true` before running commands |
| Error Handling  | HTTP 429   | Implement **exponential backoff** retry strategy              |
| Account Funding | Min $500   | Check for data permission errors if the balance is too low    |

---

