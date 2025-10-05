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

	Counter     int
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

// only should be called from OpenStream function
func (p *Pool) NextStream() int {
	if len(p.SmuxSession) == 0 {
		return -1
	}
	if p.Counter > len(p.SmuxSession)-1 {
		p.Counter = 1
		return 0
	} else {
		c := p.Counter
		p.Counter = p.Counter + 1
		return c
	}
}
func (p *Pool) OpenStream() *smux.Stream {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-time.After(3 * time.Second):
		return nil
	default:
		for {
			ns := p.NextStream()
			if ns == -1 {
				return nil
			}
			stream, err := p.SmuxSession[ns].OpenStream()
			if err == smux.ErrGoAway {
				continue
			} else if err != nil {
				go p.Remove(p.SmuxSession[ns])
				continue
			}
			return stream
		}
	}
}
