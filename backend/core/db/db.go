package db

import (
	"backend/main/config"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// initializes the database if it isn't initialized
func InitDB() {
	abs, _ := filepath.Abs("../../core/db/migrations/models.sql")
	log.Println("Attempting to read schema SQL from:", abs)
	err, isDatabaseExists := ensureDatabaseExists(config.Database_Url, config.Database_Name)
	if err != nil {
		log.Fatalf("Error checking/creating DB: %v", err)
	}

	poolCfg, err := pgxpool.ParseConfig(config.Database_Url)
	if err != nil {
		log.Fatalf("Error parsing connection string: %v", err)
	}
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	DB, err = pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := DB.Ping(ctx); err != nil {
		log.Fatalf("Database ping error: %v", err)
	}
	log.Println("Successful connection to PostgreSQL")
	if !isDatabaseExists {
		if err := InitSchemaIfNeeded("../../core/db/migrations/models.sql"); err != nil {
			log.Fatalf("Schema initialization error: %v", err)
		}
	}
}

// checks if the database exists
func ensureDatabaseExists(dbURL, dbName string) (error, bool) {
	systemURL := strings.Replace(dbURL, dbName, "postgres", 1)
	sysPool, err := pgxpool.New(context.Background(), systemURL)
	if err != nil {
		return fmt.Errorf("failed to connect to system DB: %w", err), false
	}
	defer sysPool.Close()

	var exists bool
	row := sysPool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, dbName)
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("error checking for database existence: %w", err), false
	}

	if exists {
		log.Printf("Database %q already exists", dbName)
		return nil, true
	}

	if _, err := sysPool.Exec(context.Background(),
		fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)); err != nil {
		return fmt.Errorf("failed to create DB %q: %w", dbName, err), false
	}
	log.Printf("Database %q successfuly created", dbName)
	return nil, false
}

// initializes the database schemes if they are not yet
func InitSchemaIfNeeded(pathToSQL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := DB.QueryRow(ctx, `
        SELECT EXISTS(
          SELECT FROM information_schema.tables
          WHERE table_schema='public' AND table_name='users'
        )`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking existence of users table: %w", err)
	}
	if exists {
		log.Println("Schema is already initialized (users table found).")
		return nil
	}

	log.Println("Initializing scheme:", pathToSQL)

	content, err := os.ReadFile(pathToSQL)
	if err != nil {
		return fmt.Errorf("failed to read .sql file %s: %w", pathToSQL, err)
	}
	sqlText := string(content)
	stmts := splitSQLStatements(sqlText)

	for _, stmt := range stmts {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := DB.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("error while executing SQL:\n%s\n%w", stmt, err)
		}
	}

	log.Println("Scheme has been initialized successfully..")
	return nil
}

// simple function to split sql expressions
func splitSQLStatements(sql string) []string {
	var stmts []string
	for _, part := range strings.Split(sql, ";") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			stmts = append(stmts, trimmed+";")
		}
	}
	return stmts
}
