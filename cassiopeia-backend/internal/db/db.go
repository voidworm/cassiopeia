package db

import (
	"context"
	"fmt"
	"strings"

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
			class_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name     VARCHAR(255) NOT NULL UNIQUE,
			colour   VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS investigators (
			investigator_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name            VARCHAR(255) NOT NULL UNIQUE,
			class_id        INT REFERENCES classes(class_id)
		)`,
		`CREATE TABLE IF NOT EXISTS campaigns (
			campaign_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name        VARCHAR(255) NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS scenarios (
			scenario_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name        VARCHAR(255) NOT NULL UNIQUE,
			campaign_id INT REFERENCES campaigns(campaign_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id        INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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

// SeedInitial inserts the static reference data (classes, investigators,
// campaigns, scenarios) that is identical across every environment, from the
// CSVs embedded under internal/db/seed/. It never touches
// sessions/session_players, since those are play data. Every insert is
// ON CONFLICT (name) DO NOTHING, so re-running is harmless.
func (d *DB) SeedInitial(ctx context.Context) error {
	classes, err := readSeedCSV("classes.csv")
	if err != nil {
		return err
	}
	for _, row := range classes {
		if _, err := d.pool.Exec(ctx,
			`INSERT INTO classes (name, colour) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`,
			row[0], row[1],
		); err != nil {
			return err
		}
	}

	investigators, err := readSeedCSV("investigators.csv")
	if err != nil {
		return err
	}
	for _, row := range investigators {
		if _, err := d.pool.Exec(ctx,
			`INSERT INTO investigators (name, class_id) VALUES ($1, (SELECT class_id FROM classes WHERE name = $2)) ON CONFLICT (name) DO NOTHING`,
			row[0], row[1],
		); err != nil {
			return err
		}
	}

	campaigns, err := readSeedCSV("campaigns.csv")
	if err != nil {
		return err
	}
	for _, row := range campaigns {
		if _, err := d.pool.Exec(ctx,
			`INSERT INTO campaigns (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`,
			row[0],
		); err != nil {
			return err
		}
	}

	scenarios, err := readSeedCSV("scenarios.csv")
	if err != nil {
		return err
	}
	for _, row := range scenarios {
		if _, err := d.pool.Exec(ctx,
			`INSERT INTO scenarios (name, campaign_id) VALUES ($1, (SELECT campaign_id FROM campaigns WHERE name = $2)) ON CONFLICT (name) DO NOTHING`,
			row[0], row[1],
		); err != nil {
			return err
		}
	}
	return nil
}

// SeedPersonal inserts play data (sessions/session_players) from
// internal/db/seed/personal-sessions.csv. Sessions have no natural unique
// key, so this guards against re-running by skipping entirely if the
// sessions table is already non-empty.
func (d *DB) SeedPersonal(ctx context.Context) error {
	var count int
	if err := d.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	rows, err := readSeedCSV("personal-sessions.csv")
	if err != nil {
		return err
	}
	for _, row := range rows {
		scenario, players := row[0], strings.Split(row[1], ";")

		var sessionID int
		err := d.pool.QueryRow(ctx,
			`INSERT INTO sessions (scenario_id) VALUES ((SELECT scenario_id FROM scenarios WHERE name = $1)) RETURNING session_id`,
			scenario,
		).Scan(&sessionID)
		if err != nil {
			return err
		}
		for _, player := range players {
			if _, err := d.pool.Exec(ctx,
				`INSERT INTO session_players (session_id, investigator_id) VALUES ($1, (SELECT investigator_id FROM investigators WHERE name = $2))`,
				sessionID, player,
			); err != nil {
				return err
			}
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

// CreateSession inserts a new play session with its participating investigators.
func (d *DB) CreateSession(ctx context.Context, scenarioID int, investigatorIDs []int) (*models.Session, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var session models.Session
	if err := tx.QueryRow(ctx,
		`INSERT INTO sessions (scenario_id) VALUES ($1) RETURNING session_id, scenario_id, session_timestamp`,
		scenarioID,
	).Scan(&session.ID, &session.ScenarioID, &session.Timestamp); err != nil {
		return nil, err
	}

	for _, invID := range investigatorIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_players (session_id, investigator_id) VALUES ($1, $2)`,
			session.ID, invID,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &session, nil
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
