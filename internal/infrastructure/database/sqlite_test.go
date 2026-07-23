package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/opinedajr/micro-investing/internal/shared/config"
	"github.com/opinedajr/micro-investing/internal/shared/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewSQLiteDatabase(t *testing.T) {
	t.Run("success - creates SQLite database instance", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver: "sqlite",
			Name:   "data/app.db",
		}

		sqliteDB := NewSQLiteDatabase(cfg, log)

		assert.NotNil(t, sqliteDB)
		assert.Nil(t, sqliteDB.db)
		assert.Nil(t, sqliteDB.sqlDB)
	})
}

func TestSQLiteDatabase_Connect(t *testing.T) {
	tests := []struct {
		name        string
		config      config.DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "error - path in non-existent directory",
			config: config.DatabaseConfig{
				Driver: "sqlite",
				Name:   "/nonexistent_dir_123456789/app.db",
			},
			expectError: true,
			errorMsg:    "failed to connect to database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewLogger("error")
			ctx := context.Background()
			sqliteDB := NewSQLiteDatabase(tt.config, log)
			db, err := sqliteDB.Connect(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}

	t.Run("success - connects to file database and pings", func(t *testing.T) {
		tmpDir := t.TempDir()
		log := logger.NewLogger("error")
		ctx := context.Background()
		cfg := config.DatabaseConfig{
			Driver: "sqlite",
			Name:   filepath.Join(tmpDir, "app.db"),
		}

		sqliteDB := NewSQLiteDatabase(cfg, log)

		db1, err := sqliteDB.Connect(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, db1)

		db2, err := sqliteDB.Connect(ctx)
		assert.NoError(t, err)
		assert.Equal(t, db1, db2, "connect should be idempotent")

		assert.NoError(t, sqliteDB.Ping())
		assert.NoError(t, sqliteDB.Close())
		assert.NoError(t, sqliteDB.Close(), "close after close should not error")
	})
}

func TestSQLiteDatabase_Migrate(t *testing.T) {
	t.Run("success - migrate returns nil (no-op)", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver: "sqlite",
			Name:   "data/app.db",
		}

		sqliteDB := NewSQLiteDatabase(cfg, log)

		type TestModel struct {
			ID   uint
			Name string
		}

		err := sqliteDB.Migrate(&TestModel{})

		assert.NoError(t, err)
	})
}

func TestSQLiteDatabase_Close(t *testing.T) {
	t.Run("success - close without connection returns nil", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver: "sqlite",
			Name:   "data/app.db",
		}

		sqliteDB := NewSQLiteDatabase(cfg, log)
		err := sqliteDB.Close()

		assert.NoError(t, err)
	})
}

func TestSQLiteDatabase_Ping(t *testing.T) {
	t.Run("error - ping without connection returns error", func(t *testing.T) {
		log := logger.NewLogger("error")
		cfg := config.DatabaseConfig{
			Driver: "sqlite",
			Name:   "data/app.db",
		}

		sqliteDB := NewSQLiteDatabase(cfg, log)
		err := sqliteDB.Ping()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not connected")
	})
}
