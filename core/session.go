package core

import (
	"math/rand"
	"net"
	"sync"
)

type Sessions struct {
	mu sync.Mutex
	Sc map[uint16]net.Conn
}

func (s *Sessions) New(c net.Conn) (id uint16, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for range 10 {
		rs := uint16(rand.Intn(1 << 16))
		_, ok := s.Sc[rs]
		if ok {
			continue
		}
		s.Sc[rs] = c
		return rs, true
	}
	return 0, false
}
func (s *Sessions) Add(id uint16, c net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sc[id] = c
}

func (s *Sessions) Remove(id uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.Sc[id]

	if ok {
		delete(s.Sc, id)
	}
}
func (s *Sessions) Get(id uint16) net.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.Sc[id]
	if !ok {
		return nil
	}
	return c
}
