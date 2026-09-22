package webui

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
)

func RegisterRoutes(r *gin.Engine) {
	dist, err := fs.Sub(DistFS, "dist")
	if err != nil {
		log.Println("webui dist not found, serving api only")
		return
	}

	spa := NewSPAHandler(dist)

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "NOT_FOUND",
					Message: "Route not found",
				},
			})
			return
		}
		spa.ServeHTTP(c.Writer, c.Request)
	})
}
