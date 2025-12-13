package kv

import (
	"encoding/json"
	"net/http"
	"nexus-engine/internal/pkg/logger"
	"sync"
	"time"
)

// Item — единица хранения
type Item struct {
	Value     any   `json:"value"`
	ExpiresAt int64 `json:"expires_at"`
}

// Options — настройки, передаваемые извне (из флагов CLI)
type Options struct {
	PersistPath        string
	SaveInterval       time.Duration
	CleanupInterval    time.Duration
	UpstreamURL        string
	UpstreamEnabled    bool
	DefaultUpstreamTTL int
	Logger             *logger.Logger
}

// Storage — структура модуля
type Storage struct {
	items map[string]Item
	mu    sync.RWMutex
	opts  Options
	log   *logger.Logger
}

// New создает новый инстанс KV
func New(opts Options) *Storage {
	s := &Storage{
		items: make(map[string]Item),
		opts:  opts,
		log:   opts.Logger,
	}

	// Запускаем фоновые задачи (см. worker.go)
	s.startWorkers()

	return s
}

// Set — сохранить значение
func (s *Storage) Set(key string, value any, ttlSeconds int) {
	var expires int64
	if ttlSeconds > 0 {
		expires = time.Now().Add(time.Duration(ttlSeconds) * time.Second).UnixNano()
	} else {
		// Если 0, считаем вечным (или очень долгим) для простоты
		expires = time.Now().Add(time.Hour * 24 * 365).UnixNano()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = Item{Value: value, ExpiresAt: expires}

	s.log.Debug("SET key='%s', ttl=%ds", key, ttlSeconds)
}

// Get — получить значение (с поддержкой Upstream)
func (s *Storage) Get(key string) (Item, bool) {
	s.mu.RLock()
	item, ok := s.items[key]
	s.mu.RUnlock()

	// Проверка TTL
	if ok {
		if time.Now().UnixNano() <= item.ExpiresAt {
			return item, true
		}
		// Если протухло — считаем, что нет (очистится потом)
		ok = false
	}

	// Cache-Aside (Upstream)
	if !ok && s.opts.UpstreamEnabled && s.opts.UpstreamURL != "" {
		s.log.Info("🌐 Miss! Fetching '%s' from upstream...", key)
		return s.fetchFromUpstream(key)
	}

	return Item{}, false
}

func (s *Storage) fetchFromUpstream(key string) (Item, bool) {
	start := time.Now()
	// Простой GET запрос к внешнему источнику
	resp, err := http.Get(s.opts.UpstreamURL + "/" + key)
	if err != nil || resp.StatusCode != http.StatusOK {
		s.log.Debug("Upstream error for '%s': %v", key, err)
		return Item{}, false
	}
	defer resp.Body.Close()

	var remoteValue any
	if err := json.NewDecoder(resp.Body).Decode(&remoteValue); err != nil {
		return Item{}, false
	}

	// Сохраняем в кэш
	s.Set(key, remoteValue, s.opts.DefaultUpstreamTTL)

	s.log.Debug("Upstream success for '%s' in %v", key, time.Since(start))

	// Возвращаем свежие данные
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items[key], true
}
