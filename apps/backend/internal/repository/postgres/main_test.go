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
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE
		);
	`
	if _, err := testDB.Exec(schema); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func cleanDB(t *testing.T, tables ...string) {
	for _, table := range tables {
		_, err := testDB.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE")
		assert.NoError(t, err)
	}
}
