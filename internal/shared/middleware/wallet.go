package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

func WalletMiddleware(service wallet.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		walletID := c.Param("walletId")
		if walletID == "" {
			c.JSON(http.StatusNotFound, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "WALLET_NOT_FOUND",
					Message: "Wallet not found",
				},
			})
			c.Abort()
			return
		}

		if _, err := service.Find(c.Request.Context(), walletID); err != nil {
			c.JSON(http.StatusNotFound, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "WALLET_NOT_FOUND",
					Message: "Wallet not found",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
