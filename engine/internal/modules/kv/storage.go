package kv

import (
	"encoding/json"
	"net/http"
	"nexus-engine/internal/pkg/logger"
	"os"
	"path/filepath"
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
	shards     [ShardCount]*Shard
	wal        *WAL
	opts       Options
	log        *logger.Logger
	snapshotMu sync.RWMutex
}

// LoadSnapshot загружает "базовое" состояние из JSON
func (s *Storage) LoadSnapshot(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} // Если файла нет — это норм
		return err
	}
	defer file.Close()

	var flatMap map[string]Item
	if err := json.NewDecoder(file).Decode(&flatMap); err != nil {
		return err
	}

	s.log.Debug("📦 Loading snapshot with %d keys...", len(flatMap))

	count := 0
	for k, v := range flatMap {
		// Восстанавливаем только живые ключи
		if v.ExpiresAt == 0 || time.Now().UnixNano() < v.ExpiresAt {
			s.restoreFromWAL(k, v.Value, v.ExpiresAt)
			count++
		}
	}
	s.log.Debug("📦 Loaded %d active keys from snapshot", count)
	return nil
}

// CreateSnapshot сохраняет текущее состояние и чистит WAL
func (s *Storage) CreateSnapshot() error {
	s.log.Debug("📸 Starting snapshot...")

	// 1. Блокируем запись (Set), чтобы согласовать состояние WAL и памяти
	s.snapshotMu.Lock()

	// 2. Собираем данные из всех шардов
	allItems := make(map[string]Item)
	now := time.Now().UnixNano()

	for _, shard := range s.shards {
		// RLock шардов нужен, чтобы не конфликтовать с внутренними процессами (типа Get или Cleanup)
		shard.mu.RLock()
		for k, v := range shard.items {
			if v.ExpiresAt > now {
				allItems[k] = v
			}
		}
		shard.mu.RUnlock()
	}

	// Очищаем WAL пока у нас эксклюзивный доступ (никто не пишет)
	if err := s.wal.Truncate(); err != nil {
		s.snapshotMu.Unlock() // Не забываем разлочить при ошибке
		s.log.Error("❌ Failed to truncate WAL: %v", err)
		return err
	}

	// Всё, состояние зафиксировано, WAL чист. Можно разрешить запись новым клиентам.
	s.snapshotMu.Unlock()

	// 2. Тяжелая операция записи JSON происходит в фоне, не блокируя Set/Get
	snapshotPath := s.opts.PersistPath
	tmpPath := snapshotPath + ".tmp"

	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	// Используем буферизацию для ускорения записи
	enc := json.NewEncoder(file)
	if err := enc.Encode(allItems); err != nil {
		file.Close()
		return err
	}
	file.Close()

	if err := os.Rename(tmpPath, snapshotPath); err != nil {
		return err
	}

	s.log.Info("📸 Snapshot created successfully (%d items)", len(allItems))
	return nil
}

// New создает новый инстанс KV
func New(opts Options) (*Storage, error) {
	s := &Storage{
		opts: opts,
		log:  opts.Logger,
	}

	for i := 0; i < ShardCount; i++ {
		s.shards[i] = NewShard()
	}

	walPath := opts.PersistPath + ".wal"
	snapshotPath := opts.PersistPath

	// 1. Создаем папку (обязательно перед чтением)
	dir := filepath.Dir(walPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 2. СНАЧАЛА грузим Snapshot (Базовое состояние)
	if err := s.LoadSnapshot(snapshotPath); err != nil {
		s.log.Error("Failed to load snapshot: %v", err)
		// Не критично, может файла еще нет
	} else {
		s.log.Debug("📦 Snapshot loaded")
	}

	// 3. ЗATEM накатываем WAL (Последние изменения поверх базы)
	if err := ReplayWAL(walPath, s); err != nil {
		s.log.Error("WAL Replay error: %v", err)
	}

	// 4. Открываем WAL для новых записей
	wal, err := OpenWAL(walPath)
	if err != nil {
		return nil, err
	}
	s.wal = wal
	s.log.Info("💾 Persistence enabled: %s", walPath)

	s.startWorkers()
	return s, nil
}

// restoreFromWAL — спец. метод для восстановления (принимает уже готовый timestamp)
func (s *Storage) restoreFromWAL(key string, value any, expiresAt int64) {
	// Если ключ уже протух пока сервер лежал — не загружаем его в память
	if expiresAt > 0 && time.Now().UnixNano() > expiresAt {
		return
	}

	idx := getShardIndex(key)
	shard := s.shards[idx]

	shard.mu.Lock()
	shard.items[key] = Item{Value: value, ExpiresAt: expiresAt}
	shard.mu.Unlock()
}

// Set — Публичный метод: пишет в WAL -> потом в RAM
func (s *Storage) Set(key string, value any, ttlSeconds int) {
	// Блокируем Снапшоттинг, но разрешаем другим Set работать параллельно
	s.snapshotMu.RLock()
	defer s.snapshotMu.RUnlock()

	var expires int64
	if ttlSeconds > 0 {
		expires = time.Now().Add(time.Duration(ttlSeconds) * time.Second).UnixNano()
	} else {
		expires = time.Now().Add(time.Hour * 24 * 365 * 100).UnixNano()
	}

	// 1. Пишем в WAL (атомарно внутри WAL.WriteEvent)
	if s.wal != nil {
		// Ошибки WAL логируем, но не роняем запрос (лучше потерять персистенцию, чем доступность)
		if err := s.wal.WriteEvent(WALEntry{Op: "set", Key: key, Value: value, Exp: expires}); err != nil {
			s.log.Error("WAL Write Error: %v", err)
		}
	}

	// 2. Пишем в RAM
	s.restoreFromWAL(key, value, expires)
	s.log.Debug("SET key='%s'", key)
}

// Get — получить значение
func (s *Storage) Get(key string) (Item, bool) {
	idx := getShardIndex(key)
	shard := s.shards[idx]

	shard.mu.RLock()
	item, ok := shard.items[key]
	shard.mu.RUnlock()

	// Проверка TTL (ленивое удаление не делаем, просто скрываем)
	if ok {
		if time.Now().UnixNano() <= item.ExpiresAt {
			return item, true
		}
		// Протухло — считаем что не нашли
		ok = false
	}

	// === Upstream Logic ===
	if !ok && s.opts.UpstreamEnabled && s.opts.UpstreamURL != "" {
		s.log.Info("🌐 Miss! Fetching '%s' from upstream...", key)
		return s.fetchFromUpstream(key)
	}

	return Item{}, false
}

// fetchFromUpstream делает HTTP запрос и сохраняет результат
func (s *Storage) fetchFromUpstream(key string) (Item, bool) {
	start := time.Now()

	// Используем client с таймаутом, чтобы не зависнуть
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(s.opts.UpstreamURL + "/" + key)

	if err != nil || resp.StatusCode != http.StatusOK {
		s.log.Debug("Upstream error for '%s': %v", key, err)
		return Item{}, false
	}
	defer resp.Body.Close()

	var remoteValue any
	if err := json.NewDecoder(resp.Body).Decode(&remoteValue); err != nil {
		return Item{}, false
	}

	// Сохраняем (это запишет и в WAL, и в память)
	// Используем DefaultUpstreamTTL из настроек
	s.Set(key, remoteValue, s.opts.DefaultUpstreamTTL)

	s.log.Debug("Upstream success for '%s' in %v", key, time.Since(start))

	// Возвращаем то, что только что сохранили (немного неоптимально читать снова, но надежно)
	// Либо можно сконструировать Item вручную, зная expiration
	return s.Get(key)
}

// Close закрывает файл журнала
func (s *Storage) Close() error {
	if s.wal != nil {
		return s.wal.Close()
	}
	return nil
}
