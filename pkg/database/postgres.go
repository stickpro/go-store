package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient struct {
	DB *pgxpool.Pool
}

func NewPostgresClient(ctx context.Context, dsn string, minConns int32, maxConns int32) (*PostgresClient, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	cfg.MinConns = minConns
	cfg.MaxConns = maxConns

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresClient{DB: pool}, nil
}

func (c *PostgresClient) EnsureSchemaMigrationsReady(ctx context.Context) error {
	res, err := c.DB.Query(ctx, "SELECT sm.version, sm.dirty FROM schema_migrations sm WHERE sm.dirty=true LIMIT 1;")
	if err != nil {
		return fmt.Errorf("checking schema migrations: %w", err)
	}
	defer res.Close()

	if res.Next() {
		return fmt.Errorf("database schema is dirty")
	}

	return nil
}

func (c *PostgresClient) Close() error {
	if c.DB != nil {
		c.DB.Close()
	}
	return nil
}
