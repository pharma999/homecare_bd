package middleware

import (
	"strings"

	"home_care_backend/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	origins := strings.Split(config.AppConfig.AllowedOrigins, ",")

	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: false,
	}

	if len(origins) == 1 && origins[0] == "*" {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = origins
		cfg.AllowCredentials = true
	}

	return cors.New(cfg)
}
