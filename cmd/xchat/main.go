package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	core_rabbitmq "github.com/Rics69/x-net/internal/core/broker/rabbitmq"
	core_config "github.com/Rics69/x-net/internal/core/config"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_postgres_gorm "github.com/Rics69/x-net/internal/core/repository/postgres/gorm"
	core_http_middleware "github.com/Rics69/x-net/internal/core/transport/http/middleware"
	core_http_server "github.com/Rics69/x-net/internal/core/transport/http/server"
	core_ws "github.com/Rics69/x-net/internal/core/transport/ws"
	auth_service "github.com/Rics69/x-net/internal/features/auth/service"
	auth_transport_http "github.com/Rics69/x-net/internal/features/auth/transport/http"
	posts_postgres_repository "github.com/Rics69/x-net/internal/features/posts/repository/postgres"
	posts_service "github.com/Rics69/x-net/internal/features/posts/service"
	posts_transport_http "github.com/Rics69/x-net/internal/features/posts/transport/http"
	posts_transport_rabbitmq "github.com/Rics69/x-net/internal/features/posts/transport/rabbitmq"
	posts_transport_ws "github.com/Rics69/x-net/internal/features/posts/transport/ws"
	users_postgres_repository "github.com/Rics69/x-net/internal/features/users/repository/postgres"
	users_service "github.com/Rics69/x-net/internal/features/users/service"
	users_transport_http "github.com/Rics69/x-net/internal/features/users/transport/http"
	"github.com/Rics69/x-net/web"
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

	logger.Debug("initializing websocket hub")

	wsHub := core_ws.NewHub(logger)
	wsTransportHTTP := core_ws.NewWSHTTPHandler(wsHub, authMiddleware)

	// хаб живёт столько же, сколько приложение: по ctx (SIGINT/SIGTERM) закроет всех клиентов
	go wsHub.Run(ctx)

	logger.Debug("initializing feature", zap.String("feature", "posts"))

	postsRepository := posts_postgres_repository.NewPostsRepository(db)

	// сервису всё равно, кто доставляет события - он видит только интерфейс
	var postsEventsPublisher posts_service.PostsEventsPublisher

	rabbitConfig := core_rabbitmq.NewConfigMust()
	if rabbitConfig.Enabled {
		logger.Debug("initializing rabbitmq client")

		rabbitClient := core_rabbitmq.NewClient(rabbitConfig, logger)

		postsEventsConsumer := posts_transport_rabbitmq.NewPostsEventsConsumer(wsHub, logger)
		rabbitClient.OnConnect(postsEventsConsumer.Setup)

		// не блокируемся на старте, если брокер ещё не поднялся: Run сам будет переподключаться,
		// а HTTP пока работает (события без брокера просто не уходят)
		rabbitDone := make(chan struct{})
		go func() {
			defer close(rabbitDone)
			rabbitClient.Run(ctx)
		}()

		// на выходе ждём, пока Run штатно закроет соединение. cancel() здесь же - на случай,
		// если main выходит не по сигналу (например порт занят), иначе ждали бы вечно
		defer func() {
			cancel()
			<-rabbitDone
		}()

		postsEventsPublisher = posts_transport_rabbitmq.NewPostsEventsPublisher(rabbitClient)
	} else {
		logger.Warn("rabbitmq disabled: posts events are delivered only to clients of this instance")

		postsEventsPublisher = posts_transport_ws.NewPostsEventsPublisher(wsHub)
	}

	postsService := posts_service.NewPostsService(postsRepository, postsEventsPublisher)
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
	apiVersionRouter.RegisterRoutes(wsTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouter)

	httpServer.RegisterSwagger()

	httpServer.RegisterStatic(web.Handler())

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
