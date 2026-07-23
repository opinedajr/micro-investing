package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type MemoryDatabase struct {
	t     *testing.T
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewMemoryDatabase(t *testing.T) *MemoryDatabase {
	return &MemoryDatabase{
		t: t,
	}
}

func (m *MemoryDatabase) Connect(ctx context.Context) (*gorm.DB, error) {
	if m.db != nil {
		return m.db, nil
	}

	m.t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	m.t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			m.t.Logf("warning: failed to close database connection: %v", err)
		}
	})

	m.db = db
	m.sqlDB = sqlDB

	return m.db, nil
}

func (m *MemoryDatabase) Close() error {
	if m.sqlDB != nil {
		return m.sqlDB.Close()
	}
	return nil
}

func (m *MemoryDatabase) Ping() error {
	if m.sqlDB == nil {
		return fmt.Errorf("database not connected")
	}
	return m.sqlDB.Ping()
}

func (m *MemoryDatabase) Migrate(models ...interface{}) error {
	if len(models) == 0 {
		return fmt.Errorf("no models provided for migration")
	}
	if m.db == nil {
		return fmt.Errorf("database not connected")
	}
	return m.db.AutoMigrate(models...)
}
