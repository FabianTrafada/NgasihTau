// Package postgres provides PostgreSQL implementations of the Pod Service repositories.
package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the PostgreSQL connection pool for the Pod Service.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB creates a new DB instance with the given connection pool.
func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

// Pool returns the underlying connection pool.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the database connection pool.
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}
