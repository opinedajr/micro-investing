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

	"github.com/opinedajr/micro-investing/internal/dashboard"
	"github.com/opinedajr/micro-investing/internal/di"
	"github.com/opinedajr/micro-investing/internal/dividends"
	"github.com/opinedajr/micro-investing/internal/healthcheck"
	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/position"
	"github.com/opinedajr/micro-investing/internal/stock"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

type E2ESuite struct {
	suite.Suite
	server    *httptest.Server
	expect    *httpexpect.Expect
	transport *http.Transport
	db        *sql.DB
	container *di.Container
}

func TestMain(m *testing.M) {
	// INV-14: os loops persistConn da stdlib net/http terminam de forma
	// assincrona apos o fechamento do body da resposta, mesmo com
	// DisableKeepAlives e CloseIdleConnections no teardown (Opcao B testada e
	// insuficiente: a saida dos loops corrida com o goleak). Como o ciclo de
	// vida dessas goroutines e interno a stdlib e fora do controle da suite,
	// optou-se por ignora-las no goleak (Opcao A) sem enfraquecer as
	// verificacoes das demais goroutines.
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("net/http.(*Server).Serve"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
	)
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
	position.RegisterRoutes(v1, s.container.PositionHandler(), s.container.WalletService())
	dashboard.RegisterRoutes(v1, s.container.DashboardHandler(), s.container.WalletService())
	dividends.RegisterRoutes(v1, s.container.DividendHandler(), s.container.WalletService())

	s.server = httptest.NewServer(r)

	// INV-14: mesmo com DisableKeepAlives, os loops persistConn da stdlib
	// terminam de forma assincrona apos o fechamento do body; o Transport e
	// fechado explicitamente no TearDownSuite via CloseIdleConnections para
	// reduzir as conexoes remanescentes antes do teardown do goleak.
	s.transport = &http.Transport{
		DisableKeepAlives: true,
	}

	s.expect = httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  s.server.URL,
		Reporter: httpexpect.NewAssertReporter(s.T()),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(s.T()),
		},
		Client: &http.Client{
			Transport: s.transport,
		},
	})
}

func (s *E2ESuite) TearDownSuite() {
	if s.transport != nil {
		s.transport.CloseIdleConnections()
	}
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
	err = s.container.DB().Exec("DELETE FROM dividends;").Error
	s.Require().NoError(err, "falha ao limpar tabela dividends")
	err = s.container.DB().Exec("DELETE FROM wallets;").Error
	s.Require().NoError(err, "falha ao limpar tabela wallets")
	err = s.container.DB().Exec("DELETE FROM positions;").Error
	s.Require().NoError(err, "falha ao limpar tabela positions")
	err = s.container.DB().Exec("DELETE FROM stocks_current_prices;").Error
	s.Require().NoError(err, "falha ao limpar tabela stocks_current_prices")
}
