package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"gobankapi/internal/config"
	"gobankapi/internal/router"
	"gobankapi/internal/scheduler"
)

func main() {
	config.LoadConfig()
	config.InitDB()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	r := router.SetupRouter()

	// Стартуем автоматический шедулер
	go scheduler.StartScheduler(config.DB, 12) // каждые 12 часов

	addr := fmt.Sprintf(":%s", config.AppConfig.Port)
	logrus.Infof("Сервер запущен на %s", addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		logrus.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
