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
	statements := []string{
		`CREATE TABLE IF NOT EXISTS classes (
			class_id INT PRIMARY KEY,
			name     VARCHAR(255) NOT NULL,
			colour   VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS investigators (
			investigator_id INT PRIMARY KEY,
			name            VARCHAR(255) NOT NULL,
			class_id        INT REFERENCES classes(class_id)
		)`,
		`CREATE TABLE IF NOT EXISTS campaigns (
			campaign_id INT PRIMARY KEY,
			name        VARCHAR(255) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS scenarios (
			scenario_id INT PRIMARY KEY,
			name        VARCHAR(255) NOT NULL,
			campaign_id INT REFERENCES campaigns(campaign_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id        INT PRIMARY KEY,
			scenario_id       INT REFERENCES scenarios(scenario_id),
			session_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS session_players (
			session_id      INT REFERENCES sessions(session_id),
			investigator_id INT REFERENCES investigators(investigator_id),
			PRIMARY KEY (session_id, investigator_id)
		)`,
	}
	for _, stmt := range statements {
		if _, err := d.pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// PreseedIfEmpty inserts a small, plain set of rows across all tables so the
// app has something to show. This is a placeholder until proper seed jobs
// are in place.
func (d *DB) PreseedIfEmpty(ctx context.Context) error {
	var count int
	if err := d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM investigators`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	statements := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO classes (class_id, name, colour) VALUES ($1, $2, $3)`, []any{1, "Guardian", "#2b6cb0"}},
		{`INSERT INTO classes (class_id, name, colour) VALUES ($1, $2, $3)`, []any{2, "Seeker", "#d69e2e"}},

		{`INSERT INTO investigators (investigator_id, name, class_id) VALUES ($1, $2, $3)`, []any{1, "Daisy Walker", 2}},
		{`INSERT INTO investigators (investigator_id, name, class_id) VALUES ($1, $2, $3)`, []any{2, "Roland Banks", 1}},
		{`INSERT INTO investigators (investigator_id, name, class_id) VALUES ($1, $2, $3)`, []any{3, "Mark Harrigan", 1}},

		{`INSERT INTO campaigns (campaign_id, name) VALUES ($1, $2)`, []any{1, "Night of the Zealot"}},

		{`INSERT INTO scenarios (scenario_id, name, campaign_id) VALUES ($1, $2, $3)`, []any{1, "The Gathering", 1}},
		{`INSERT INTO scenarios (scenario_id, name, campaign_id) VALUES ($1, $2, $3)`, []any{2, "The Midnight Masks", 1}},

		{`INSERT INTO sessions (session_id, scenario_id) VALUES ($1, $2)`, []any{1, 1}},
		{`INSERT INTO sessions (session_id, scenario_id) VALUES ($1, $2)`, []any{2, 2}},

		{`INSERT INTO session_players (session_id, investigator_id) VALUES ($1, $2)`, []any{1, 1}},
		{`INSERT INTO session_players (session_id, investigator_id) VALUES ($1, $2)`, []any{1, 2}},
		{`INSERT INTO session_players (session_id, investigator_id) VALUES ($1, $2)`, []any{2, 1}},
		{`INSERT INTO session_players (session_id, investigator_id) VALUES ($1, $2)`, []any{2, 3}},
	}
	for _, s := range statements {
		if _, err := d.pool.Exec(ctx, s.sql, s.args...); err != nil {
			return err
		}
	}
	return nil
}

func scanInvestigator(row interface {
	Scan(dest ...any) error
}) (*models.Investigator, error) {
	var inv models.Investigator
	if err := row.Scan(&inv.ID, &inv.Name, &inv.ClassID, &inv.PlayCount); err != nil {
		return nil, err
	}
	return &inv, nil
}

func investigatorSelect(where string) string {
	return fmt.Sprintf(`
		SELECT i.investigator_id, i.name, i.class_id, COUNT(sp.session_id)
		FROM investigators i
		LEFT JOIN session_players sp ON sp.investigator_id = i.investigator_id
		%s
		GROUP BY i.investigator_id, i.name, i.class_id
	`, where)
}

func (d *DB) ListInvestigators(ctx context.Context) ([]models.Investigator, error) {
	rows, err := d.pool.Query(ctx, investigatorSelect("")+` ORDER BY COUNT(sp.session_id) DESC, i.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Investigator
	for rows.Next() {
		inv, err := scanInvestigator(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *inv)
	}
	return result, rows.Err()
}

func (d *DB) GetInvestigatorByID(ctx context.Context, id int) (*models.Investigator, error) {
	row := d.pool.QueryRow(ctx, investigatorSelect(`WHERE i.investigator_id = $1`), id)
	return scanInvestigator(row)
}

func (d *DB) GetInvestigatorByName(ctx context.Context, name string) (*models.Investigator, error) {
	row := d.pool.QueryRow(ctx, investigatorSelect(`WHERE i.name = $1`), name)
	return scanInvestigator(row)
}

func (d *DB) ListClasses(ctx context.Context) ([]models.Class, error) {
	rows, err := d.pool.Query(ctx, `SELECT class_id, name, colour FROM classes ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.Name, &c.Colour); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (d *DB) ListCampaigns(ctx context.Context) ([]models.Campaign, error) {
	rows, err := d.pool.Query(ctx, `SELECT campaign_id, name FROM campaigns ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (d *DB) ListScenarios(ctx context.Context) ([]models.Scenario, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT s.scenario_id, s.name, s.campaign_id, COUNT(se.session_id)
		FROM scenarios s
		LEFT JOIN sessions se ON se.scenario_id = s.scenario_id
		GROUP BY s.scenario_id, s.name, s.campaign_id
		ORDER BY COUNT(se.session_id) DESC, s.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Scenario
	for rows.Next() {
		var sc models.Scenario
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.CampaignID, &sc.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, sc)
	}
	return result, rows.Err()
}

// CountPlayedTogether returns how many sessions two investigators have both played in.
func (d *DB) CountPlayedTogether(ctx context.Context, investigatorAID, investigatorBID int) (int, error) {
	var count int
	err := d.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM session_players sp1
		JOIN session_players sp2 ON sp1.session_id = sp2.session_id
		WHERE sp1.investigator_id = $1 AND sp2.investigator_id = $2
	`, investigatorAID, investigatorBID).Scan(&count)
	return count, err
}

// CountPlaysByClass returns how many session_players rows belong to investigators of the given class.
func (d *DB) CountPlaysByClass(ctx context.Context, classID int) (int, error) {
	var count int
	err := d.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM session_players sp
		JOIN investigators i ON i.investigator_id = sp.investigator_id
		WHERE i.class_id = $1
	`, classID).Scan(&count)
	return count, err
}
