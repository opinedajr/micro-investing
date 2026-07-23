package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/di"
)

func main() {
	container := di.NewContainer()
	port := container.Config().Server.Port
	r := gin.Default()

	r.GET("/health", container.HealthCheckHandler().Handle)

	log.Fatal(r.Run(":" + port))
}
