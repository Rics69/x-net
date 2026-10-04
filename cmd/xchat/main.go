package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	core_config "github.com/Rics69/x-net/internal/core/config"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_postgres_gorm "github.com/Rics69/x-net/internal/core/repository/postgres/gorm"
	core_http_middleware "github.com/Rics69/x-net/internal/core/transport/http/middleware"
	core_http_server "github.com/Rics69/x-net/internal/core/transport/http/server"
	auth_service "github.com/Rics69/x-net/internal/features/auth/service"
	auth_transport_http "github.com/Rics69/x-net/internal/features/auth/transport/http"
	posts_postgres_repository "github.com/Rics69/x-net/internal/features/posts/repository/postgres"
	posts_service "github.com/Rics69/x-net/internal/features/posts/service"
	posts_transport_http "github.com/Rics69/x-net/internal/features/posts/transport/http"
	users_postgres_repository "github.com/Rics69/x-net/internal/features/users/repository/postgres"
	users_service "github.com/Rics69/x-net/internal/features/users/service"
	users_transport_http "github.com/Rics69/x-net/internal/features/users/transport/http"
	"go.uber.org/zap"

	// docs генерится через `make swagger-gen`, в init() регистрирует спеку в swag,
	// откуда её потом читает httpSwagger.Handler
	_ "github.com/Rics69/x-net/docs"
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

	logger.Debug("initializing postgres GORM connection")
	db, err := core_postgres_gorm.NewDB(ctx, core_postgres_gorm.NewConfigMust(), logger)
	if err != nil {
		logger.Fatal("failed to init postgres GORM connection", zap.Error(err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close postgres GORM connection", zap.Error(err))
		}
	}()

	logger.Debug("initializing auth token manager")

	authConfig := core_auth.NewConfigMust()
	tokenManager, err := core_auth.NewTokenManager(authConfig)
	if err != nil {
		logger.Fatal("failed to init auth token manager", zap.Error(err))
	}

	authMiddleware := core_http_middleware.Auth(tokenManager)

	logger.Debug("initializing feature", zap.String("feature", "users"))

	usersRepository := users_postgres_repository.NewUsersRepository(db)
	usersService := users_service.NewUsersService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService, authMiddleware)

	logger.Debug("initializing feature", zap.String("feature", "auth"))

	authService := auth_service.NewAuthService(usersRepository, tokenManager)
	authTransportHTTP := auth_transport_http.NewAuthHTTPHandler(authService, authConfig.CookieSecure)

	logger.Debug("initializing feature", zap.String("feature", "posts"))

	postsRepository := posts_postgres_repository.NewPostsRepository(db)
	postsService := posts_service.NewPostsService(postsRepository)
	postsTransportHTTP := posts_transport_http.NewPostsHTTPHandler(postsService, authMiddleware)

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
	apiVersionRouter.RegisterRoutes(authTransportHTTP.Routes()...)
	apiVersionRouter.RegisterRoutes(postsTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouter)

	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
