package kv

import (
	"flag"
	"net/http"
	"nexus-engine/internal/core"
	"nexus-engine/internal/pkg/logger"
	"time"
)

// Убеждаемся, что Module реализует интерфейс core.Module
var _ core.Module = (*Module)(nil)

type Module struct {
	store *Storage

	// Флаги CLI
	fDataDir         *string
	fUpstreamURL     *string
	fUpstreamTTL     *int
	fSaveInterval    *int
	fCleanupInterval *int
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "KV-Store"
}

func (m *Module) RegisterFlags(fs *flag.FlagSet) {
	m.fDataDir = fs.String("kv-data-dir", "./data", "Directory for KV persistence")

	// Интервалы в секундах
	m.fSaveInterval = fs.Int("kv-save-interval", 30, "Interval in seconds to save to disk")
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

	var err error
	m.store, err = New(opts)
	if err != nil {
		return err
	}

	return nil
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/kv/get", m.handleGet)
	mux.HandleFunc("/kv/set", m.handleSet)
}

func (m *Module) Shutdown() {
	if m.store != nil {
		m.store.log.Info("Stopping KV Store...")
		m.store.CreateSnapshot()
		m.store.Close()
	}
}
