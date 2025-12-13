package kv

import (
	"flag"
	"fmt"
	"net/http"
	"nexus-engine/internal/core"
	"nexus-engine/internal/pkg/logger"
	"os"
	"time"
)

// Убеждаемся, что Module реализует интерфейс core.Module
var _ core.Module = (*Module)(nil)

type Module struct {
	store   *Storage
	options Options // Теперь мы будем это использовать

	// Флаги CLI
	fDataDir         *string
	fUpstreamURL     *string
	fUpstreamTTL     *int
	fSaveInterval    *int // <--- Добавили флаг
	fCleanupInterval *int // <--- Добавили флаг
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "KV-Store"
}

func (m *Module) RegisterFlags(fs *flag.FlagSet) {
	m.fDataDir = fs.String("kv-data-dir", "./nexus-data", "Directory for KV persistence")

	// Интервалы в секундах
	m.fSaveInterval = fs.Int("kv-save-interval", 5, "Interval in seconds to save to disk")
	m.fCleanupInterval = fs.Int("kv-cleanup-interval", 10, "Interval in seconds to remove expired keys")

	m.fUpstreamURL = fs.String("kv-upstream-url", "", "URL for cache-aside pattern")
	m.fUpstreamTTL = fs.Int("kv-upstream-ttl", 60, "TTL for upstream items")
}

func (m *Module) Init(log *logger.Logger) error {
	// Собираем конфиг из флагов
	opts := Options{
		PersistPath: *m.fDataDir + "/kv.json",
		// Используем значения из флагов
		SaveInterval:    time.Duration(*m.fSaveInterval) * time.Second,
		CleanupInterval: time.Duration(*m.fCleanupInterval) * time.Second,

		UpstreamURL:        *m.fUpstreamURL,
		UpstreamEnabled:    *m.fUpstreamURL != "",
		DefaultUpstreamTTL: *m.fUpstreamTTL,
		Logger:             log,
	}

	// 1. Исправляем "Unused variable": Сохраняем конфиг в структуру модуля
	m.options = opts

	// 2. Инициализируем хранилище
	m.store = New(m.options)

	// Создаем директорию для хранения данных, если ее нет
	if err := os.MkdirAll(*m.fDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory '%s': %w", *m.fDataDir, err)
	}

	if err := m.store.LoadFromFile(); err != nil {
		log.Info("[%s] Starting fresh (no dump found)", m.Name())
	}

	return nil
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/kv/get", m.handleGet)
	mux.HandleFunc("/kv/set", m.handleSet)
}

func (m *Module) Shutdown() {
	if m.store != nil {
		m.store.SaveToFile()
	}
}
