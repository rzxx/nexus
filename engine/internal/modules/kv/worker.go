package kv

import (
	"time"
)

func (s *Storage) startWorkers() {
	go func() {
		ticker := time.NewTicker(s.opts.CleanupInterval)
		for range ticker.C {
			s.cleanupShards()
		}
	}()

	// Snapshot Ticker (например, раз в 1 минуту или 5 минут)
	go func() {
		if s.opts.SaveInterval <= 0 {
			s.log.Info("Snapshot interval is 0, auto-save disabled")
			return
		}
		ticker := time.NewTicker(s.opts.SaveInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := s.CreateSnapshot(); err != nil {
				s.log.Error("Snapshot failed: %v", err)
			}
		}
	}()
}

func (s *Storage) cleanupShards() {
	// Проходимся по каждому шарду по очереди, чтобы не грузить CPU
	for i := 0; i < ShardCount; i++ {
		s.cleanupSingleShard(s.shards[i])
	}
}

func (s *Storage) cleanupSingleShard(shard *Shard) {
	sampleSize := 20
	now := time.Now().UnixNano()

	shard.mu.Lock()
	defer shard.mu.Unlock()

	processed := 0
	for key, item := range shard.items {
		if processed >= sampleSize {
			break
		}
		if now > item.ExpiresAt {
			delete(shard.items, key)
			// В идеале: записать в WAL событие {"op":"del", "k":key}
			// Но для TTL это не обязательно, при перезагрузке они и так будут старыми
		}
		processed++
	}
}
