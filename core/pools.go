package core

import (
	"sync"
	"time"

	"github.com/xtaci/smux"
)

type Pool struct {
	mu sync.Mutex

	Size int
	Tls  bool
	Addr string

	SmuxSession []*smux.Session
}

func NewPool(max_size int) *Pool {
	p := Pool{Size: max_size}
	return &p
}

func (p *Pool) Add(s *smux.Session) {

	p.mu.Lock()
	defer p.mu.Unlock()
	p.SmuxSession = append(p.SmuxSession, s)

}

func (p *Pool) Remove(session *smux.Session) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, s := range p.SmuxSession {
		if s == session {
			p.SmuxSession = append(p.SmuxSession[:i], p.SmuxSession[i+1:]...)
			session.Close()
			return true
		}
	}
	return false
}

func (p *Pool) OpenStream() *smux.Stream {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.SmuxSession) == 0 {
		return nil
	}
	select {
	case <-time.After(5 * time.Second):
		return nil
	default:
		for _, s := range p.SmuxSession {
			stream, err := s.OpenStream()
			if err == smux.ErrGoAway {
				continue
			} else if err != nil {
				go p.Remove(s)
				continue
			}
			return stream
		}

		return nil
	}
}
