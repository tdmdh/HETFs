package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdmdh/HETFs/internal/domain"
)

// ETFRepository defines the persistence contract for mapping records.
type ETFRepository interface {
	// UpsertContract caches a contract mapping.
	UpsertContract(ctx context.Context, contract domain.Contract) error
	// UpsertContracts bulk-caches mappings.
	UpsertContracts(ctx context.Context, contracts []domain.Contract) error
	// GetContractBySymbol finds a cached contract.
	GetContractBySymbol(ctx context.Context, symbol string) (*domain.Contract, error)
	// ListAllContracts returns every cached contract.
	ListAllContracts(ctx context.Context) ([]domain.Contract, error)
}

// sqliteETFRepo implements ETFRepository backed by SQLite.
type sqliteETFRepo struct {
	db *sql.DB
}

// NewETFRepository creates an ETFRepository backed by the given SQLite DB.
func NewETFRepository(db *sql.DB) ETFRepository {
	return &sqliteETFRepo{db: db}
}

func (r *sqliteETFRepo) UpsertContract(ctx context.Context, contract domain.Contract) error {
	const query = `
INSERT INTO contracts (conid, symbol, company_name, exchange)
VALUES (?, ?, ?, ?)
ON CONFLICT(conid) DO UPDATE SET
    symbol       = excluded.symbol,
    company_name = excluded.company_name,
    exchange     = excluded.exchange;`

	_, err := r.db.ExecContext(ctx, query, contract.ConID, contract.Symbol, contract.CompanyName, contract.Exchange)
	if err != nil {
		return fmt.Errorf("upsert contract %v: %w", contract.ConID, err)
	}
	return nil
}

func (r *sqliteETFRepo) UpsertContracts(ctx context.Context, contracts []domain.Contract) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const query = `
INSERT INTO contracts (conid, symbol, company_name, exchange)
VALUES (?, ?, ?, ?)
ON CONFLICT(conid) DO UPDATE SET
    symbol       = excluded.symbol,
    company_name = excluded.company_name,
    exchange     = excluded.exchange;`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare upsert: %w", err)
	}
	defer stmt.Close()

	for _, c := range contracts {
		if _, err := stmt.ExecContext(ctx, c.ConID, c.Symbol, c.CompanyName, c.Exchange); err != nil {
			return fmt.Errorf("upsert contract %v: %w", c.ConID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *sqliteETFRepo) GetContractBySymbol(ctx context.Context, symbol string) (*domain.Contract, error) {
	const query = `SELECT id, conid, symbol, company_name, exchange FROM contracts WHERE symbol = ? LIMIT 1;`
	row := r.db.QueryRowContext(ctx, query, symbol)

	var c domain.Contract
	if err := row.Scan(&c.ID, &c.ConID, &c.Symbol, &c.CompanyName, &c.Exchange); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("scan contract: %w", err)
	}
	return &c, nil
}

func (r *sqliteETFRepo) ListAllContracts(ctx context.Context) ([]domain.Contract, error) {
	const query = `SELECT id, conid, symbol, company_name, exchange FROM contracts ORDER BY symbol;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query contracts: %w", err)
	}
	defer rows.Close()

	var contracts []domain.Contract
	for rows.Next() {
		var c domain.Contract
		if err := rows.Scan(&c.ID, &c.ConID, &c.Symbol, &c.CompanyName, &c.Exchange); err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		contracts = append(contracts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return contracts, nil
}
