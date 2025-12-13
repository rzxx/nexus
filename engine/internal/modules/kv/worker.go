package kv

import (
	"encoding/json"
	"os"
	"time"
)

func (s *Storage) startWorkers() {
	// Воркер очистки (TTL)
	go func() {
		ticker := time.NewTicker(s.opts.CleanupInterval)
		for range ticker.C {
			s.cleanupProbabilistic()
		}
	}()

	// Воркер персистенции
	if s.opts.PersistPath != "" && s.opts.SaveInterval > 0 {
		go func() {
			ticker := time.NewTicker(s.opts.SaveInterval)
			for range ticker.C {
				s.SaveToFile()
			}
		}()
	}
}

// Алгоритм вероятностной очистки
func (s *Storage) cleanupProbabilistic() {
	sampleSize := 20
	threshold := 25 // 25%

	totalDeleted := 0
	loops := 0

	for {
		expiredCount := 0
		processedCount := 0
		now := time.Now().UnixNano()

		s.mu.Lock()
		for key, item := range s.items {
			if processedCount >= sampleSize {
				break
			}
			if now > item.ExpiresAt {
				delete(s.items, key)
				expiredCount++
			}
			processedCount++
		}
		s.mu.Unlock()

		totalDeleted += expiredCount
		loops++

		if processedCount < sampleSize || (expiredCount*100/sampleSize) <= threshold {
			break
		}
	}

	if totalDeleted > 0 {
		s.log.Info("🧹 Janitor: removed %d keys (in %d loops)", totalDeleted, loops)
	}
}

// Сохранение дампа
func (s *Storage) SaveToFile() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, err := os.Create(s.opts.PersistPath)
	if err != nil {
		s.log.Error("❌ Error creating dump: %v", err)
		return err
	}
	defer file.Close()

	s.log.Debug("💾 Saving snapshot to '%s'...", s.opts.PersistPath)
	return json.NewEncoder(file).Encode(s.items)
}

// Загрузка дампа
func (s *Storage) LoadFromFile() error {
	file, err := os.Open(s.opts.PersistPath)
	if err != nil {
		if os.IsNotExist(err) {
			s.log.Info("📂 No dump found, starting fresh.")
			return nil
		}
		return err
	}
	defer file.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("📂 Loading snapshot from '%s'...", s.opts.PersistPath)
	return json.NewDecoder(file).Decode(&s.items)
}
