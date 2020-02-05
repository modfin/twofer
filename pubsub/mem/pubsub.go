package mem

import (
	"context"
	"sync"
	"time"
)

type PubSub struct {
	mutex sync.Mutex

	counter   int64
	observers map[string][]observer
}

type observer struct {
	id      int64
	c       chan []byte
	created time.Time
}

func New() *PubSub {
	return &PubSub{
		observers: map[string][]observer{},
	}
}

func (p *PubSub) Notify(topic string, payload []byte) error {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// TODO maybe do this async?
	for _, obs := range p.observers[topic]{
		obs.c <- payload
	}

	return nil
}
func (p *PubSub) Listen(ctx context.Context, topic string) (reader <-chan []byte, cancel func()) {

	c := make(chan []byte, 1)
	p.mutex.Lock()
	defer p.mutex.Unlock()

	id := p.counter
	p.counter++
	p.observers[topic] = append(p.observers[topic], observer{
		id:      id,
		c:       c,
		created: time.Now(),
	})

	cancel = func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		obs := p.observers[topic]
		for i, o := range obs {
			if o.id == id {
				p.observers[topic] = append(obs[:i], obs[i+1:]...)
				return
			}
		}
	}
	return c, cancel
}
func (p *PubSub) Next(ctx context.Context, topic string) ([]byte, error) {
	c, cancel := p.Listen(ctx, topic)
	defer cancel()

	select {
	case data := <-c:
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
