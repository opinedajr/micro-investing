package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
)

//go:embed all:dist
var distFS embed.FS

func RegisterRoutes(r *gin.Engine) {
	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "NOT_FOUND",
					Message: "Route not found",
				},
			})
			return
		}

		name := strings.TrimPrefix(path, "/")
		if info, statErr := fs.Stat(dist, name); statErr == nil && !info.IsDir() {
			c.FileFromFS(path, http.FS(dist))
			return
		}

		index, readErr := fs.ReadFile(dist, "index.html")
		if readErr != nil {
			c.String(http.StatusNotFound, "frontend build not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}
