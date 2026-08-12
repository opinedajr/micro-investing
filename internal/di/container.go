package di

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/opinedajr/micro-investing/internal/healthcheck"
	"github.com/opinedajr/micro-investing/internal/infrastructure/database"
	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/shared/config"
	sloglogger "github.com/opinedajr/micro-investing/internal/shared/logger"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

type Container struct {
	config       *config.Config
	logger       sloglogger.Logger
	db           *gorm.DB
	dbConn       database.DatabaseConnection
	repositories *RepositoryDependencies
	services     *ServiceDependencies
	handlers     *HandlerDependencies
}

type RepositoryDependencies struct {
	walletRepository    wallet.Repository
	patrimonyRepository patrimony.PatrimonyRepository
}

type HandlerDependencies struct {
	healthcheckHandler  *healthcheck.Handler
	walletHandler       *wallet.Handler
	patrimonyHandler    *patrimony.Handler
}

type ServiceDependencies struct {
	healthcheckService *healthcheck.Service
	walletService      wallet.Service
	patrimonyService   patrimony.Service
}

func NewContainer() *Container {
	return &Container{
		repositories: &RepositoryDependencies{},
		services:     &ServiceDependencies{},
		handlers:     &HandlerDependencies{},
	}
}

func (c *Container) Config() *config.Config {
	if c.config == nil {
		cfg, err := config.Load()
		if err != nil {
			panic("failed to load config: " + err.Error())
		}
		c.config = cfg
	}
	return c.config
}

func (c *Container) Logger() sloglogger.Logger {
	if c.logger == nil {
		cfg := c.Config()
		c.logger = sloglogger.NewLogger(cfg.Logging.Level)
	}
	return c.logger
}

func (c *Container) DB() *gorm.DB {
	if c.db == nil {
		ctx := context.Background()
		var dbConn database.DatabaseConnection

		switch c.Config().Database.Driver {
		case "mysql":
			dbConn = database.NewMySQLDatabase(c.Config().Database, c.Logger())
		case "postgres":
			dbConn = database.NewPostgresDatabase(c.Config().Database, c.Logger())
		case "sqlite":
			dbConn = database.NewSQLiteDatabase(c.Config().Database, c.Logger())
		default:
			panic(fmt.Sprintf("unsupported database driver: %s", c.Config().Database.Driver))
		}

		db, err := dbConn.Connect(ctx)
		if err != nil {
			panic("failed to connect to database: " + err.Error())
		}

		c.db = db
		c.dbConn = dbConn
	}
	return c.db
}

func (c *Container) HealthCheckService() *healthcheck.Service {
	if c.services.healthcheckService == nil {
		_ = c.DB()
		c.services.healthcheckService = healthcheck.NewHealthCheckService(c.dbConn)
	}
	return c.services.healthcheckService
}

func (c *Container) HealthCheckHandler() *healthcheck.Handler {
	if c.handlers.healthcheckHandler == nil {
		c.handlers.healthcheckHandler = healthcheck.NewHandler(c.HealthCheckService())
	}
	return c.handlers.healthcheckHandler
}

func (c *Container) WalletRepository() wallet.Repository {
	if c.repositories.walletRepository == nil {
		c.repositories.walletRepository = wallet.NewSQLiteRepository(c.DB())
	}
	return c.repositories.walletRepository
}

func (c *Container) WalletService() wallet.Service {
	if c.services.walletService == nil {
		c.services.walletService = wallet.NewService(c.WalletRepository())
	}
	return c.services.walletService
}

func (c *Container) WalletHandler() *wallet.Handler {
	if c.handlers.walletHandler == nil {
		c.handlers.walletHandler = wallet.NewHandler(c.WalletService())
	}
	return c.handlers.walletHandler
}

func (c *Container) PatrimonyRepository() patrimony.PatrimonyRepository {
	if c.repositories.patrimonyRepository == nil {
		c.repositories.patrimonyRepository = patrimony.NewSQLiteRepository(c.DB())
	}
	return c.repositories.patrimonyRepository
}

func (c *Container) PatrimonyService() patrimony.Service {
	if c.services.patrimonyService == nil {
		c.services.patrimonyService = patrimony.NewService(c.PatrimonyRepository())
	}
	return c.services.patrimonyService
}

func (c *Container) PatrimonyHandler() *patrimony.Handler {
	if c.handlers.patrimonyHandler == nil {
		c.handlers.patrimonyHandler = patrimony.NewHandler(c.PatrimonyService())
	}
	return c.handlers.patrimonyHandler
}
