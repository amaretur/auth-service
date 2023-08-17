package app

import (
	"time"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/amaretur/auth-service/config"

	"github.com/amaretur/auth-service/pkg/log"

	"github.com/amaretur/auth-service/internal/infrastructure/server"

	http "github.com/amaretur/auth-service/internal/transport/http/handler"
	"github.com/amaretur/auth-service/internal/usecase"
	"github.com/amaretur/auth-service/internal/service"
	"github.com/amaretur/auth-service/internal/repository"
)

type App struct {
	config		*config.Config
	logger		log.Logger

	httpHandler	*http.Handler

	httpServer	*server.Http

	onClearFuncs []func()
}

func NewApp(conf *config.Config, logger log.Logger) *App {
	return &App{
		config: conf,
		logger: logger,
	}
}

func (a *App) Init() error {

	// Установка соединения с mongodb
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.
		Client().
		ApplyURI(a.config.MongoDB.ConnectURL()).
		SetServerAPIOptions(serverAPI)

	ctx1, cancel1 := context.WithTimeout(
		context.Background(),
		a.config.MongoDB.OpenTimeout*time.Second,
	)
	defer cancel1()

	client, err := mongo.Connect(ctx1, opts)

	if err != nil {
		a.logger.Errorf("connect to mongodb: %s", err)

		return err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	err = client.
		Database(a.config.MongoDB.Database).
		RunCommand(ctx2, bson.D{{"ping", 1}}).
		Err()

	if err != nil {
		a.logger.Errorf("ping mongodb connect: %s", err)

		return err
	}

	// Закрытие соединения с mongodb
	a.onClear(func() {

		ctx, cancel := context.
			WithTimeout(context.Background(), 20*time.Second)

		defer cancel()

		if err := client.Disconnect(ctx); err != nil {
			a.logger.Errorf("close mongodb connect: %s", err)
		}

		a.logger.Info("connection to mongodb successfully closed")
	})

	// Создание репозитория
	repo := repository.NewTokenRepositoryMongo(
		client.Database(a.config.MongoDB.Database),
		a.logger.WithFields(map[string]any{"layer": "repository"}),
	)

	// Создание сервисов
	jwtService := service.NewJwt(
		repo,
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

func (a *App) onClear(f func()) {
	a.onClearFuncs = append(a.onClearFuncs, f)
}

func (a *App) clear() {

	for _, f := range a.onClearFuncs {
		f()
	}
}


