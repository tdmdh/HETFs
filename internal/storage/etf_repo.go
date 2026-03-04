package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdmdh/HETFs/internal/domain"
)

// ETFRepository defines the persistence contract for ETF records.
type ETFRepository interface {
	// UpsertETF inserts or replaces a single ETF.
	UpsertETF(ctx context.Context, etf domain.ETF) error
	// UpsertETFs inserts or replaces multiple ETFs in a single transaction.
	UpsertETFs(ctx context.Context, etfs []domain.ETF) error
	// ListETFs returns ETFs filtered by exchange.
	ListETFs(ctx context.Context, exchange string) ([]domain.ETF, error)
	// ListAllETFs returns every stored ETF.
	ListAllETFs(ctx context.Context) ([]domain.ETF, error)
}

// sqliteETFRepo implements ETFRepository backed by SQLite.
type sqliteETFRepo struct {
	db *sql.DB
}

// NewETFRepository creates an ETFRepository backed by the given SQLite DB.
func NewETFRepository(db *sql.DB) ETFRepository {
	return &sqliteETFRepo{db: db}
}

func (r *sqliteETFRepo) UpsertETF(ctx context.Context, etf domain.ETF) error {
	const query = `
INSERT INTO etfs (symbol, isin, exchange, currency)
VALUES (?, ?, ?, ?)
ON CONFLICT(isin, exchange) DO UPDATE SET
    symbol   = excluded.symbol,
    currency = excluded.currency;`

	_, err := r.db.ExecContext(ctx, query, etf.Symbol, etf.ISIN, etf.Exchange, etf.Currency)
	if err != nil {
		return fmt.Errorf("upsert etf %s: %w", etf.Symbol, err)
	}
	return nil
}

func (r *sqliteETFRepo) UpsertETFs(ctx context.Context, etfs []domain.ETF) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const query = `
INSERT INTO etfs (symbol, isin, exchange, currency)
VALUES (?, ?, ?, ?)
ON CONFLICT(isin, exchange) DO UPDATE SET
    symbol   = excluded.symbol,
    currency = excluded.currency;`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare upsert: %w", err)
	}
	defer stmt.Close()

	for _, etf := range etfs {
		if _, err := stmt.ExecContext(ctx, etf.Symbol, etf.ISIN, etf.Exchange, etf.Currency); err != nil {
			return fmt.Errorf("upsert etf %s: %w", etf.Symbol, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *sqliteETFRepo) ListETFs(ctx context.Context, exchange string) ([]domain.ETF, error) {
	const query = `SELECT id, symbol, isin, exchange, currency FROM etfs WHERE exchange = ? ORDER BY symbol;`
	return r.queryETFs(ctx, query, exchange)
}

func (r *sqliteETFRepo) ListAllETFs(ctx context.Context) ([]domain.ETF, error) {
	const query = `SELECT id, symbol, isin, exchange, currency FROM etfs ORDER BY exchange, symbol;`
	return r.queryETFs(ctx, query)
}

func (r *sqliteETFRepo) queryETFs(ctx context.Context, query string, args ...any) ([]domain.ETF, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query etfs: %w", err)
	}
	defer rows.Close()

	var etfs []domain.ETF
	for rows.Next() {
		var e domain.ETF
		if err := rows.Scan(&e.ID, &e.Symbol, &e.ISIN, &e.Exchange, &e.Currency); err != nil {
			return nil, fmt.Errorf("scan etf: %w", err)
		}
		etfs = append(etfs, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return etfs, nil
}
