// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"tarkib.uz/config"
	v1 "tarkib.uz/internal/controller/http/v1"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/internal/usecase/repo"
	"tarkib.uz/internal/usecase/webapi"
	"tarkib.uz/pkg/httpserver"
	"tarkib.uz/pkg/logger"
	"tarkib.uz/pkg/postgres"
	"tarkib.uz/pkg/redis"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	RedisClient, err := redis.NewRedisDB(cfg)
	if err != nil {
		pp.Println(err)
		l.Fatal(fmt.Errorf("app - Run - redis.New: %w", err))
	}

	endpoint := os.Getenv("SERVER_IP")
	accessKeyID := "nodirbek"
	secretAccessKey := "nodirbek"
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - minio.New"))
	}

	// Use case
	authUseCase := usecase.NewAuthUseCase(
		repo.NewAuthRepo(pg),
		webapi.NewAuthWebAPI(cfg),
		cfg,
		RedisClient,
		minioClient,
	)

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, l, authUseCase)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
