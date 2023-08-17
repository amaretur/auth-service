package app

import (
	"github.com/amaretur/auth-service/config"

	"github.com/amaretur/auth-service/pkg/log"

	"github.com/amaretur/auth-service/internal/infrastructure/server"

	http "github.com/amaretur/auth-service/internal/transport/http/handler"
	"github.com/amaretur/auth-service/internal/usecase"
	"github.com/amaretur/auth-service/internal/service"
)

type App struct {
	config		*config.Config
	logger		log.Logger

	httpHandler	*http.Handler

	httpServer	*server.Http
}

func NewApp(conf *config.Config, logger log.Logger) *App {
	return &App{
		config: conf,
		logger: logger,
	}
}

func (a *App) Init() error {

	// Создание сервисов
	jwtService := service.NewJwt(
		a.config.Jwt.AccessExpire,
		a.config.Jwt.RefreshExpire,
		a.config.Jwt.Secret,
		a.logger.WithFields(map[string]any{"layer": "service"}),
	)

	// Создание юзкейсов
	authUsecase := usecase.New(
		jwtService,
		a.logger.WithFields(map[string]any{"layer": "usecase"}),
	)

	// Логгер для обработчиков
	httpLogger := a.logger.WithFields(map[string]any{
		"layer": "transport",
		"protocol": "http",
	})

	// Создание и регистрация обработчиков
	handler := http.NewHandler("/api/v1")

	handler.Register(http.NewAuth(authUsecase, httpLogger), "")

	a.httpHandler = handler

	return nil
}

func (a *App) Run() error {

	a.httpServer = server.NewHttp(a.logger)

	errChan := make(chan error, 1)

	// Запуск HTTP сервера
	go a.httpServer.Run(errChan, &server.HttpConfig{
		Port:			a.config.Http.Port,
		MaxHeaderBytes:	a.config.Http.MaxHeaderBytes,
		ReadTimeout:	a.config.Http.ReadTimeout,
		WriteTimeout:	a.config.Http.WriteTimeout,
		Handler:		a.httpHandler.Router(),
	})

	if err := <- errChan; err != nil {
		a.logger.Error(err)

		a.clear()

		return err
	}

	return nil
}

func (a *App) Shutdown() {

	a.logger.Info("stopping the application...")

	if err := a.httpServer.Shutdown(); err != nil {
		a.logger.Error(err)
	}

	a.logger.Info("http server is stopped")

	a.clear()
}

func (a *App) clear() {
}


