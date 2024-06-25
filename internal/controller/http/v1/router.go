// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"net/http"

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
// @description tarkib.uz 
// @version     1.0
// @host        8080-idx-go-clean-template-1719253883593.cluster-blu4edcrfnajktuztkjzgyxzek.cloudworkstations.dev
// @BasePath    /v1
func NewRouter(handler *gin.Engine, l logger.Interface, t usecase.Auth) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// handler.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost", "http://anotherdomain.com"}, // Update with your allowed origins
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))

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
	}
}
