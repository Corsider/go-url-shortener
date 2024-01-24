package inmemory

import (
	"sync"
)

type MemoryStorage struct {
	mu sync.RWMutex
	db map[int]string
}

func New() *MemoryStorage {
	return &MemoryStorage{
		db: make(map[int]string),
	}
}

func (m *MemoryStorage) Save(original string, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db[id] = original
	return nil
}

func (m *MemoryStorage) Load(id int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db[id], nil
}

func (m *MemoryStorage) GetLastId() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.db), nil
}

func (m *MemoryStorage) CheckExistence(original string) (bool, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for key, val := range m.db {
		if val == original {
			return true, key
		}
	}
	return false, 0
}

func (m *MemoryStorage) Close() error {
	return nil
}
