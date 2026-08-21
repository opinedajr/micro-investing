package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/di"
	"github.com/opinedajr/micro-investing/internal/healthcheck"
	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/stock"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

func main() {
	container := di.NewContainer()
	port := container.Config().Server.Port
	r := gin.Default()

	v1 := r.Group("/api/v1")
	healthcheck.RegisterRoutes(v1, container.HealthCheckHandler())
	wallet.RegisterRoutes(v1, container.WalletHandler())
	patrimony.RegisterRoutes(v1, container.PatrimonyHandler(), container.WalletService())
	stock.RegisterRoutes(v1, container.StockHandler())

	log.Fatal(r.Run(":" + port))
}
