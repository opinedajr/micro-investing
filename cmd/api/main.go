package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/di"
	"github.com/opinedajr/micro-investing/internal/healthcheck"
)

func main() {
	container := di.NewContainer()
	port := container.Config().Server.Port
	r := gin.Default()

	v1 := r.Group("/api/v1")
	healthcheck.RegisterRoutes(v1, container.HealthCheckHandler())

	log.Fatal(r.Run(":" + port))
}
