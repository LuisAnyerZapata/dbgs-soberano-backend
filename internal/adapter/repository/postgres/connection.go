package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"DBGS_SOBERANO_BACKEND/config"

	_ "github.com/lib/pq"
)

// NewPostgresConnection inicializa la conexión con la base de datos PostgreSQL utilizando la configuración recibida
func NewPostgresConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la conexión con postgres: %w", err)
	}

	db.SetMaxOpenConns(int(cfg.MaxConns))
	db.SetMaxIdleConns(int(cfg.MinConns))
	db.SetConnMaxLifetime(15 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error al verificar la conexión (ping) con postgres: %w", err)
	}

	return db, nil
}