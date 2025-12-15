package kv

import (
	"encoding/json"
	"os"
	"sync"
)

// WALEntry — одна операция в журнале
type WALEntry struct {
	Op    string `json:"op"` // "set", "del"
	Key   string `json:"k"`
	Value any    `json:"v"`
	Exp   int64  `json:"e,omitempty"`
}

type WAL struct {
	file *os.File
	path string
	mu   sync.Mutex // Блокировка только на запись в файл
	enc  *json.Encoder
}

func OpenWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
		path: path,
		enc:  json.NewEncoder(file),
	}, nil
}

// WriteEvent записывает событие в конец файла
func (w *WAL) WriteEvent(entry WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.enc.Encode(entry)
}

func (w *WAL) Close() error {
	return w.file.Close()
}

// Truncate полностью очищает файл лога
func (w *WAL) Truncate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Сначала ЗАКРЫВАЕМ файл.
	// Это обязательно для Windows, чтобы снять блокировку "Access is denied".
	if err := w.file.Close(); err != nil {
		return err
	}

	// 2. Теперь, когда файл закрыт, обрезаем его до 0 байт по пути.
	if err := os.Truncate(w.path, 0); err != nil {
		// Если не вышло обрезать, пытаемся хотя бы переоткрыть, чтобы не крашнуть систему
		_ = w.reopen()
		return err
	}

	// 3. Открываем файл заново для записи
	return w.reopen()
}

// Вспомогательный метод для открытия файла
func (w *WAL) reopen() error {
	file, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = file
	w.enc = json.NewEncoder(w.file) // Создаем новый энкодер для нового файла
	return nil
}

// Replay считывает лог и применяет его к хранилищу (используется при старте)
func ReplayWAL(path string, store *Storage) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Файла нет — база пустая, это ок
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry WALEntry
		if err := decoder.Decode(&entry); err != nil {
			return err // Битая запись
		}

		if entry.Op == "set" {
			store.restoreFromWAL(entry.Key, entry.Value, entry.Exp)
		}
	}
	return nil
}
