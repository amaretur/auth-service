package main

import (
	"os"
	"flag"
	"context"
	"syscall"

	"github.com/amaretur/auth-service/app"
	"github.com/amaretur/auth-service/config"

	"github.com/amaretur/auth-service/pkg/log"
)

func main() {

	var pathConf string

	flag.StringVar(
		&pathConf,
		"config",
		"config/config.toml",
		"The path to the configuration file",
	)

	flag.Parse()

	logger := log.NewLogrusLogger()

	// Инициализация конфигураций
	conf, err := config.Init(pathConf)
	if err != nil {
		logger.Errorf("init config: %s\n", err.Error())
		return
	}

	// Экземпляр приложения
	a := app.NewApp(conf, logger)

	// Инициализация приложения
	if err := a.Init(); err != nil {
		logger.Errorf("init error: %s\n", err.Error())
		return
	}

	// Запуск приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context, cancel context.CancelFunc) {
		if err := a.Run(); err != nil {
			logger.Errorf("run error: ", err.Error())
			cancel() // Завершаем контекст инициируя завершение прогарммы
		}
	}(ctx, cancel)

	// По указанным сигналам - завершаем работу приложения
	// Остановка приложения и высвобождение ресурсов так же будут инициированы
	// если контекс будет прерван (контекст прерывается при неудачном запуске)
	app.OnSignal(
		ctx,
		a.Shutdown,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, 
		syscall.SIGABRT, os.Interrupt,
	)
}
