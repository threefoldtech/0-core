package plugin

import "sync"

type store struct {
	data map[string][]byte
	m    sync.RWMutex
}

func newStore() *store {
	return &store{
		data: make(map[string][]byte),
	}
}

func (s *store) Set(key string, value []byte) {
	s.m.Lock()
	defer s.m.Unlock()

	s.data[key] = value
}

func (s *store) Get(key string) ([]byte, bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	data, ok := s.data[key]
	return data, ok
}

func (s *store) Del(key string) {
	s.m.RLock()
	defer s.m.RUnlock()

	delete(s.data, key)
}

func (s *store) List() map[string][]byte {
	data := make(map[string][]byte)
	s.m.RLock()
	defer s.m.RUnlock()
	for k, v := range s.data {
		data[k] = v
	}

	return data
}
