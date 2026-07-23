package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/opinedajr/micro-investing/internal/shared/config"
	"github.com/opinedajr/micro-investing/internal/shared/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type SQLiteDatabase struct {
	cfg   config.DatabaseConfig
	log   logger.Logger
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewSQLiteDatabase(cfg config.DatabaseConfig, log logger.Logger) *SQLiteDatabase {
	return &SQLiteDatabase{
		cfg: cfg,
		log: log,
	}
}

func (s *SQLiteDatabase) Connect(ctx context.Context) (*gorm.DB, error) {
	if s.db != nil {
		return s.db, nil
	}

	dsn := fmt.Sprintf(
		"%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)",
		s.cfg.Name,
	)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		s.log.Error(ctx, "failed to connect to database",
			"database", s.cfg.Name,
			"error", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.Ping(); err != nil {
		s.log.Error(ctx, "failed to ping database",
			"error", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s.log.Info(ctx, "database connection established",
		"database", s.cfg.Name)

	s.db = db
	s.sqlDB = sqlDB

	return s.db, nil
}

func (s *SQLiteDatabase) Close() error {
	if s.sqlDB != nil {
		return s.sqlDB.Close()
	}
	return nil
}

func (s *SQLiteDatabase) Ping() error {
	if s.sqlDB == nil {
		return fmt.Errorf("database not connected")
	}
	return s.sqlDB.Ping()
}

func (s *SQLiteDatabase) Migrate(models ...interface{}) error {
	return nil
}
