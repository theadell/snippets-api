package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"snippets.adelh.dev/app/internal/config"
	"snippets.adelh.dev/app/internal/db/gen"
)

// Store provides database access with primary/replica support
type Store struct {
	primary    *sql.DB
	replicas   []*sql.DB
	mu         sync.RWMutex
	q          gen.Queries
	replicaIdx int
}

func NewStore(cfg config.DBConfig) (*Store, error) {
	// Connect to primary
	primary, err := sql.Open("pgx", cfg.PrimaryDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to primary DB: %w", err)
	}

	// Configure connection pool settings for primary
	primary.SetMaxOpenConns(cfg.MaxOpenConns)
	primary.SetMaxIdleConns(cfg.MaxIdleConns)
	primary.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test primary connection
	if err := primary.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping primary DB: %w", err)
	}

	// Connect to replicas
	var replicas []*sql.DB
	for i, dsn := range cfg.ReplicaDSNs {
		replica, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to replica DB %d: %w", i, err)
		}

		// Configure connection pool settings for replica
		replica.SetMaxOpenConns(cfg.MaxOpenConns)
		replica.SetMaxIdleConns(cfg.MaxIdleConns)
		replica.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		// Test replica connection
		if err := replica.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping replica DB %d: %w", i, err)
		}

		replicas = append(replicas, replica)
	}

	// Create the store
	store := &Store{
		primary:  primary,
		replicas: replicas,
		q:        *gen.New(primary),
	}

	return store, nil
}

func (s *Store) Primary() *gen.Queries {
	return &s.q
}

// Replica returns a queries struct connected to a replica database
// Uses round-robin selection
func (s *Store) Replica() *gen.Queries {
	if len(s.replicas) == 0 {
		// Fall back to primary if no replicas
		return gen.New(s.primary)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.replicaIdx
	s.replicaIdx = (s.replicaIdx + 1) % len(s.replicas)

	return gen.New(s.replicas[idx])
}

// WithTx executes a function within a database transaction
func (s *Store) WithTx(ctx context.Context, fn func(*gen.Queries) error) error {
	tx, err := s.primary.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := gen.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
