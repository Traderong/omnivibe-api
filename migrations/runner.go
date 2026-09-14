package migrations

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var migrationFiles embed.FS

type Migration struct {
	Version int
	Name    string
	SQL     string
}

func Run(ctx context.Context, db *pgxpool.Pool) error {
	if _, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	currentVersion, err := currentVersion(ctx, db)
	if err != nil {
		return err
	}

	// The existing OmniVibe database already contains migrations 001-005,
	// but it was created before schema_migrations existed. Verify the
	// expected schema and establish 005 as the baseline.
	if currentVersion == 0 {
		initialized, err := databaseAlreadyAtBaseline(ctx, db)
		if err != nil {
			return err
		}

		if initialized {
			if err := recordBaseline(ctx, db, migrations); err != nil {
				return err
			}

			currentVersion = 5
		}
	}

	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		if err := applyMigration(ctx, db, migration); err != nil {
			return err
		}
	}

	return nil
}

func loadMigrations() ([]Migration, error) {
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		sql, err := migrationFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf(
				"read migration %s: %w",
				entry.Name(),
				err,
			)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    entry.Name(),
			SQL:     string(sql),
		})
	}

	sort.Slice(
		migrations,
		func(i, j int) bool {
			return migrations[i].Version < migrations[j].Version
		},
	)

	return migrations, nil
}

func currentVersion(
	ctx context.Context,
	db *pgxpool.Pool,
) (int, error) {
	var version int

	err := db.QueryRow(
		ctx,
		`SELECT COALESCE(MAX(version), 0)
		 FROM schema_migrations`,
	).Scan(&version)

	if err != nil {
		return 0, fmt.Errorf("get migration version: %w", err)
	}

	return version, nil
}

func databaseAlreadyAtBaseline(
	ctx context.Context,
	db *pgxpool.Pool,
) (bool, error) {
	var usersExists bool
	if err := db.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = 'users'
		)`,
	).Scan(&usersExists); err != nil {
		return false, fmt.Errorf("check users table: %w", err)
	}

	if !usersExists {
		return false, nil
	}

	var refreshTokensExists bool
	if err := db.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = 'refresh_tokens'
		)`,
	).Scan(&refreshTokensExists); err != nil {
		return false, fmt.Errorf(
			"check refresh_tokens table: %w",
			err,
		)
	}

	if !refreshTokensExists {
		return false, nil
	}

	var passwordResetTokenExists bool
	if err := db.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'users'
			  AND column_name = 'password_reset_token'
		)`,
	).Scan(&passwordResetTokenExists); err != nil {
		return false, fmt.Errorf(
			"check password reset column: %w",
			err,
		)
	}

	if !passwordResetTokenExists {
		return false, nil
	}

	var activeIndexExists bool
	if err := db.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = 'public'
			  AND tablename = 'refresh_tokens'
			  AND indexname = 'idx_refresh_tokens_active_user'
		)`,
	).Scan(&activeIndexExists); err != nil {
		return false, fmt.Errorf(
			"check active refresh token index: %w",
			err,
		)
	}

	return activeIndexExists, nil
}

func recordBaseline(
	ctx context.Context,
	db *pgxpool.Pool,
	migrations []Migration,
) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration baseline: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, migration := range migrations {
		if migration.Version > 5 {
			break
		}

		if _, err := tx.Exec(
			ctx,
			`INSERT INTO schema_migrations (version, name)
			 VALUES ($1, $2)
			 ON CONFLICT (version) DO NOTHING`,
			migration.Version,
			migration.Name,
		); err != nil {
			return fmt.Errorf(
				"record baseline migration %03d: %w",
				migration.Version,
				err,
			)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit migration baseline: %w",
			err,
		)
	}

	return nil
}

func applyMigration(
	ctx context.Context,
	db *pgxpool.Pool,
	migration Migration,
) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin migration %03d: %w",
			migration.Version,
			err,
		)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, migration.SQL); err != nil {
		return fmt.Errorf(
			"apply migration %03d (%s): %w",
			migration.Version,
			migration.Name,
			err,
		)
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO schema_migrations (version, name)
		 VALUES ($1, $2)`,
		migration.Version,
		migration.Name,
	); err != nil {
		return fmt.Errorf(
			"record migration %03d: %w",
			migration.Version,
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit migration %03d: %w",
			migration.Version,
			err,
		)
	}

	return nil
}
