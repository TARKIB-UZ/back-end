// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	_ "tarkib.uz/docs"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       tarkib.uz back-end
// @description Backend team - Nodirbek, Dostonbek
// @version     1.0
// @BasePath    /v1
// @security    BearerAuth
func NewRouter(handler *gin.Engine, l logger.Interface, t usecase.Auth) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.AllowBrowserExtensions = true
	corsConfig.AllowMethods = []string{"*"}
	handler.Use(cors.New(corsConfig))

	// Swagger
	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	handler.GET("/swagger/*any", swaggerHandler)

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newAuthRoutes(h, t, l)
		newFileRoutes(h, l)
	}
}
