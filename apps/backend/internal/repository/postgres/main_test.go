package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func TestMain(m *testing.M) {

	var (
		ctx       = context.Background()
		container testcontainers.Container
		err       error
	)
	container, testDB, err = setupPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to setup postgres container: %v", err)
	}

	setupSchema()
	code := m.Run()

	if err := container.Terminate(ctx); err != nil {
		log.Fatalf("Failed to terminate container: %v", err)
	}

	os.Exit(code)
}

func setupPostgresContainer(ctx context.Context) (testcontainers.Container, *sql.DB, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:18.1",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}

	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return container, db, nil
}

func setupSchema() {
	const schema = `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(26) PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL DEFAULT '',
			role VARCHAR(255) DEFAULT 'User',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Function to automatically update updated_at timestamp
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		-- Trigger for users table
		DROP TRIGGER IF EXISTS update_users_updated_at ON users;
		CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();

		-- Games table (needed for match foreign key)
		CREATE TABLE IF NOT EXISTS games (
			id VARCHAR(26) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			author_id VARCHAR(26) REFERENCES users(id) ON DELETE SET NULL,
			status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
			is_public BOOLEAN DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Matches table
		CREATE TABLE IF NOT EXISTS matches (
			id VARCHAR(26) PRIMARY KEY,
			user_id VARCHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			game_id VARCHAR(26) NOT NULL REFERENCES games(id) ON DELETE CASCADE,
			status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'won', 'lost', 'resigned', 'expired', 'error')),
			total_tokens INTEGER NOT NULL DEFAULT 0,
			turn_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := testDB.Exec(schema); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}
}

func cleanDB(t *testing.T, tables ...string) {
	for _, table := range tables {
		_, err := testDB.Exec("TRUNCATE TABLE " + table + " CASCADE")
		assert.NoError(t, err)
	}
}
