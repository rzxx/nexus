package kv

import (
	"hash/fnv"
	"sync"
)

const ShardCount = 32

// Shard — это маленькое независимое хранилище
type Shard struct {
	mu    sync.RWMutex
	items map[string]Item
}

func NewShard() *Shard {
	return &Shard{
		items: make(map[string]Item),
	}
}

// getShardIndex вычисляет, в каком шарде лежит ключ
func getShardIndex(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % ShardCount
}
