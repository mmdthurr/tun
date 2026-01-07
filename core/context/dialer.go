package context

import (
	"time"
	"tun/core"
	"tun/core/proto"
)

type DialerEvent struct {
	Event int
}

type DialerContext struct {
	DialPool    int
	Dialer      proto.Proto
	Handler     core.Handler
	TriggerChan chan DialerEvent
}

// listen on trigger ensure there is only one dial at each 500 Millisecond
// so no more Excessive dial will occur
func (dctx DialerContext) ListenOnTrigger() {
	for {
		ev := <-dctx.TriggerChan
		switch ev.Event {
		case 0:
			{
				go func() {
					dctx.Dialer.StartDialer(dctx.Handler)
					go dctx.Trigger()
				}()
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func (dctx DialerContext) Trigger() {

	dctx.TriggerChan <- DialerEvent{Event: 0}

}

func (dctx DialerContext) Start() {

	go dctx.ListenOnTrigger()

	for range dctx.DialPool {
		go dctx.Trigger()
	}

}
