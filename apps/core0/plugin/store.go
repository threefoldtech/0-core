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

func (s *store) Set(key string, value []byte) error {
	s.m.Lock()
	defer s.m.Unlock()

	s.data[key] = value
	return nil
}

func (s *store) Get(key string) ([]byte, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	data, _ := s.data[key]
	return data, nil
}

func (s *store) List() (map[string][]byte, error) {
	data := make(map[string][]byte)
	s.m.RLock()
	defer s.m.RUnlock()
	for k, v := range s.data {
		data[k] = v
	}

	return data, nil
}
