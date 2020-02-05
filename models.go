package twofer

import "context"

type PubSub interface {
	Notify(topic string, payload []byte) error
	Listen(ctx context.Context, topic string) (reader <-chan []byte, cancel func())
	Next(ctx context.Context, topic string) ([]byte, error)
}
