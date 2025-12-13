package core

import (
	"flag"
	"net/http"
	"nexus-engine/internal/pkg/logger"
)

// Module описывает любой подключаемый компонент (KV, Queue, etc)
type Module interface {
	// Name возвращает имя модуля для логов
	Name() string

	// RegisterFlags позволяет модулю добавить свои аргументы CLI
	RegisterFlags(fs *flag.FlagSet)

	// Init запускает логику модуля (после парсинга флагов)
	Init(log *logger.Logger) error

	// RegisterRoutes добавляет эндпоинты в общий роутер
	RegisterRoutes(mux *http.ServeMux)

	// Shutdown вызывается при остановке сервера
	Shutdown()
}
