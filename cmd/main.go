package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"simple-service/internal/api"
	"simple-service/internal/config"
	customLogger "simple-service/internal/logger"
	"simple-service/internal/repo"
	"simple-service/internal/service"
)

func main() {
	// Загружаем конфигурацию из переменных окружения

	if err := godotenv.Load("local.env"); err != nil {
		log.Fatal("Ошибка загрузки env файла:", err)
	}

	var cfg config.AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "failed to load configuration"))
	}

	// Инициализация логгера
	logger, err := customLogger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error initializing logger"))
	}

	// Подключение к PostgreSQL
	repository, err := repo.NewRepository(context.Background(), cfg.PostgreSQL)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize repository"))
	}

	// Создание сервиса с бизнес-логикой
	serviceInstance := service.NewService(repository, logger)

	// Инициализация API
	app := api.NewRouters(&api.Routers{Service: serviceInstance}, cfg.Rest.Token)

	// Запуск HTTP-сервера в отдельной горутине
	go func() {
		logger.Infof("Starting server on %s", cfg.Rest.ListenAddress)
		if err := app.Listen(cfg.Rest.ListenAddress); err != nil {
			log.Fatal(errors.Wrap(err, "failed to start server"))
		}
	}()

	// Ожидание системных сигналов для корректного завершения работы
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutting down gracefully...")
}
