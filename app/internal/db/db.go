package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"snippets.adelh.dev/app/internal/config"
	"snippets.adelh.dev/app/internal/db/sqlc"
)

type Store interface {
	Primary() sqlc.Querier
	Replica() sqlc.Querier
	WithTx(ctx context.Context, fn func(sqlc.Querier) error) error
	Close() error
}

type PostgresStore struct {
	primary    *sql.DB
	replicas   []*sql.DB
	mu         sync.RWMutex
	q          sqlc.Queries
	replicaIdx int
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(cfg config.DBConfig) (*PostgresStore, error) {
	primary, err := sql.Open("pgx", cfg.PrimaryDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to primary DB: %w", err)
	}

	primary.SetMaxOpenConns(cfg.MaxOpenConns)
	primary.SetMaxIdleConns(cfg.MaxIdleConns)
	primary.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := primary.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping primary DB: %w", err)
	}

	var replicas []*sql.DB
	for i, dsn := range cfg.ReplicaDSNs {
		replica, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to replica DB %d: %w", i, err)
		}

		replica.SetMaxOpenConns(cfg.MaxOpenConns)
		replica.SetMaxIdleConns(cfg.MaxIdleConns)
		replica.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		if err := replica.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping replica DB %d: %w", i, err)
		}

		replicas = append(replicas, replica)
	}

	store := &PostgresStore{
		primary:  primary,
		replicas: replicas,
		q:        *sqlc.New(primary),
	}

	return store, nil
}

func (s *PostgresStore) Primary() sqlc.Querier {
	return &s.q
}

// Replica returns a queries struct connected to a replica database
// Uses round-robin selection
func (s *PostgresStore) Replica() sqlc.Querier {
	if len(s.replicas) == 0 {
		return sqlc.New(s.primary)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.replicaIdx
	s.replicaIdx = (s.replicaIdx + 1) % len(s.replicas)

	return sqlc.New(s.replicas[idx])
}

// WithTx executes a function within a database transaction
func (s *PostgresStore) WithTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	tx, err := s.primary.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := sqlc.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (s *PostgresStore) Close() error {
	if err := s.primary.Close(); err != nil {
		return err
	}
	for i, replica := range s.replicas {
		if err := replica.Close(); err != nil {
			return fmt.Errorf("failed to close replica %d: %w", i, err)
		}
	}
	return nil
}
