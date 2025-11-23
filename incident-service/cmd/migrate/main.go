package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq"
	"github.com/your-org/ai-sre-platform/incident-service/internal/config"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Override with TEST_ environment variables if present
	if testHost := os.Getenv("TEST_DATABASE_HOST"); testHost != "" {
		cfg.Database.Host = testHost
	}
	if testPort := os.Getenv("TEST_DATABASE_PORT"); testPort != "" {
		_, _ = fmt.Sscanf(testPort, "%d", &cfg.Database.Port)
	}
	if testDB := os.Getenv("TEST_DATABASE_NAME"); testDB != "" {
		cfg.Database.Database = testDB
	}
	if testUser := os.Getenv("TEST_DATABASE_USER"); testUser != "" {
		cfg.Database.User = testUser
	}
	if testPassword := os.Getenv("TEST_DATABASE_PASSWORD"); testPassword != "" {
		cfg.Database.Password = testPassword
	}
	if testSSLMode := os.Getenv("TEST_DATABASE_SSL_MODE"); testSSLMode != "" {
		cfg.Database.SSLMode = testSSLMode
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.Database.DatabaseDSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("migrations completed successfully")
}

func runMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationsDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	sort.Strings(files)

	// Apply each migration
	for _, file := range files {
		version := filepath.Base(file)

		// Check if migration already applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if exists {
			fmt.Printf("skipping migration %s (already applied)\n", version)
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration
		fmt.Printf("applying migration %s...\n", version)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		fmt.Printf("migration %s applied successfully\n", version)
	}

	return nil
}
