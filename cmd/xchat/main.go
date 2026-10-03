package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/Rics69/x-chat/internal/core/config"
	core_logger "github.com/Rics69/x-chat/internal/core/logger"
	core_pgx_pool "github.com/Rics69/x-chat/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/Rics69/x-chat/internal/core/transport/http/middleware"
	core_http_server "github.com/Rics69/x-chat/internal/core/transport/http/server"
	users_postgres_repository "github.com/Rics69/x-chat/internal/features/users/repository/postgres"
	users_service "github.com/Rics69/x-chat/internal/features/users/service"
	users_transport_http "github.com/Rics69/x-chat/internal/features/users/transport/http"
	"go.uber.org/zap"

	// docs генерится через `make swagger-gen`, в init() регистрирует спеку в swag,
	// откуда её потом читает httpSwagger.Handler
	_ "github.com/Rics69/x-chat/docs"
)

// @title X-Chat API
// @version 1.0
// @description X-Chat App REST-API schema
// @host 127.0.0.1:5050
// @BasePath /api/v1
func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)

	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}

	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", time.Local))

	logger.Debug("initializing postgres connection pool")
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}

	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))

	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing HTTP server")

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.CORS(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	apiVersionRouter := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(usersTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouter)

	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
