//go:build integration

package e2e

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/opinedajr/micro-investing/internal/di"
	"github.com/opinedajr/micro-investing/internal/healthcheck"
	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/stock"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

type E2ESuite struct {
	suite.Suite
	server    *httptest.Server
	expect    *httpexpect.Expect
	db        *sql.DB
	container *di.Container
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("net/http.(*Server).Serve"))
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}

func (s *E2ESuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)

	tempDir, err := os.MkdirTemp("", "micro-investing-e2e-*")
	s.Require().NoError(err)
	dbPath := filepath.Join(tempDir, "test.db")

	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_NAME", dbPath)
	os.Setenv("SERVER_PORT", "0")

	s.container = di.NewContainer()

	db, err := s.container.DB().DB()
	s.Require().NoError(err, "failed to get underlying sql.DB")
	s.db = db

	migrationsPath := "file://../../migrations"
	dbURL := "sqlite3://" + dbPath

	m, err := migrate.New(migrationsPath, dbURL)
	s.Require().NoError(err, "failed to initialize migrator")

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		s.Require().NoError(err, "failed to run migrations")
	}

	_, _ = m.Close()

	err = s.container.StockService().Seed(context.Background(), stock.SeedInput{Force: false})
	s.Require().NoError(err, "failed to seed stocks")

	r := gin.Default()
	v1 := r.Group("/api/v1")
	healthcheck.RegisterRoutes(v1, s.container.HealthCheckHandler())
	wallet.RegisterRoutes(v1, s.container.WalletHandler())
	patrimony.RegisterRoutes(v1, s.container.PatrimonyHandler(), s.container.WalletService())
	stock.RegisterRoutes(v1, s.container.StockHandler())

	s.server = httptest.NewServer(r)

	s.expect = httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  s.server.URL,
		Reporter: httpexpect.NewAssertReporter(s.T()),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(s.T()),
		},
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	})
}

func (s *E2ESuite) TearDownSuite() {
	s.server.Close()
	if s.db != nil {
		s.db.Close()
	}
}

func (s *E2ESuite) SetupTest() {
	err := s.container.DB().Exec("DELETE FROM assets;").Error
	s.Require().NoError(err, "falha ao limpar tabela assets")
	err = s.container.DB().Exec("DELETE FROM patrimonies;").Error
	s.Require().NoError(err, "falha ao limpar tabela patrimonies")
	err = s.container.DB().Exec("DELETE FROM wallets;").Error
	s.Require().NoError(err, "falha ao limpar tabela wallets")
}
