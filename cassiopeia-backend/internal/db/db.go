package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fogshaper/cassiopeia-backend/internal/models"
)

type DB struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (d *DB) Close() {
	d.pool.Close()
}

func (d *DB) Migrate(ctx context.Context) error {
	_, err := d.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS investigators (
			uid        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name       TEXT NOT NULL UNIQUE,
			play_count INTEGER NOT NULL DEFAULT 0
		)
	`)
	return err
}

func (d *DB) SeedIfEmpty(ctx context.Context) error {
	var count int
	if err := d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM investigators`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	seed := []struct {
		name      string
		playCount int
	}{
		{"Daisy Walker", 101},
		{"Norman Withers", 33},
		{"Daniela Reyes", 113},
		{"Mark Harrigan", 8},
	}
	for _, s := range seed {
		if _, err := d.pool.Exec(ctx,
			`INSERT INTO investigators (name, play_count) VALUES ($1, $2)`,
			s.name, s.playCount,
		); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) ListInvestigators(ctx context.Context) ([]models.Investigator, error) {
	rows, err := d.pool.Query(ctx, `SELECT uid, name, play_count FROM investigators ORDER BY play_count DESC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Investigator
	for rows.Next() {
		var inv models.Investigator
		if err := rows.Scan(&inv.UID, &inv.Name, &inv.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, inv)
	}
	return result, rows.Err()
}

func (d *DB) GetInvestigatorByID(ctx context.Context, uid string) (*models.Investigator, error) {
	var inv models.Investigator
	err := d.pool.QueryRow(ctx,
		`SELECT uid, name, play_count FROM investigators WHERE uid = $1`, uid,
	).Scan(&inv.UID, &inv.Name, &inv.PlayCount)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (d *DB) GetInvestigatorByName(ctx context.Context, name string) (*models.Investigator, error) {
	var inv models.Investigator
	err := d.pool.QueryRow(ctx,
		`SELECT uid, name, play_count FROM investigators WHERE name = $1`, name,
	).Scan(&inv.UID, &inv.Name, &inv.PlayCount)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (d *DB) IncrementPlayCount(ctx context.Context, uid string, by int) (*models.Investigator, error) {
	var inv models.Investigator
	err := d.pool.QueryRow(ctx,
		`UPDATE investigators SET play_count = play_count + $2 WHERE uid = $1 RETURNING uid, name, play_count`,
		uid, by,
	).Scan(&inv.UID, &inv.Name, &inv.PlayCount)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (d *DB) SetPlayCount(ctx context.Context, uid string, value int) (*models.Investigator, error) {
	var inv models.Investigator
	err := d.pool.QueryRow(ctx,
		`UPDATE investigators SET play_count = $2 WHERE uid = $1 RETURNING uid, name, play_count`,
		uid, value,
	).Scan(&inv.UID, &inv.Name, &inv.PlayCount)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}
