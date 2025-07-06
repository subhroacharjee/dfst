package p2p

import (
	"context"
	"net"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
)

type Peer interface {
	net.Conn
	SetID(string)
	GetID() string
	Close() error
	Send([]byte) error
}

type Transport interface {
	ID() string
	ListenAndAccept(context.Context) error
	Subsribe() broadcaster.Subsriber
	Shutdown(context.CancelFunc)
	Dial(context.Context, string) error
	Broadcast([]byte)
	Send(string, []byte) error
}

type TransportOpts struct {
	PeerID        string
	ListenerAddr  string
	Decoder       Decoder
	HandShakeFunc HandShakeFunc
	PeerAddr      []string
}
