package p2p

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
	"github.com/subhroacharjee/dfst/internal/logger"
)

type TCPTransportOpts struct {
	PeerID        string
	ListenerAddr  string
	Decoder       Decoder
	HandShakeFunc HandShakeFunc
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu      sync.RWMutex
	msgChan *broadcaster.Broadcaster

	peers map[string]Peer
}

func NewTCPTransport(opts TCPTransportOpts) Transport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		peers:            make(map[string]Peer),
		msgChan:          broadcaster.NewBroadcaster(context.Background()),
	}
}

func (t *TCPTransport) ID() string {
	return t.PeerID
}

func (t *TCPTransport) ListenAndAccept(ctx context.Context) (err error) {
	ln, err := net.Listen("tcp", t.TCPTransportOpts.ListenerAddr)
	if err != nil {
		return err
	}

	t.listener = ln
	go t.startAcceptLoop(ctx)
	return nil
}

func (t *TCPTransport) Dial(ctx context.Context, addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logger.Error("Dialing to %s has failed", addr)
		return err
	}

	go func() {
		logger.Debug("New connection dialed")
		t.handleConnection(ctx, conn, true)
	}()

	return nil
}

func (t *TCPTransport) startAcceptLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("Closing accept loop")
			return
		default:
			conn, err := t.listener.Accept()
			if err != nil {
				logger.Info("TCP accept closed")
				return
			}

			go func() {
				logger.Debug("new connection made")
				t.handleConnection(ctx, conn, false)
			}()
		}
	}
}

func (t *TCPTransport) handleConnection(ctx context.Context, conn net.Conn, outbound bool) {
	logger.Info("Connected")
	peer := NewTCPPeer(conn, outbound)

	if err := t.TCPTransportOpts.HandShakeFunc(peer); err != nil {
		logger.Error("Handshake failed with new connection %s", peer.Conn.RemoteAddr().String())
		peer.Close()
		return
	}

	t.mu.Lock()
	t.peers[peer.ID] = peer
	t.mu.Unlock()

	msg := broadcaster.Message{
		From: peer.GetID(),
	}
	for {
		select {
		case <-ctx.Done():
			if err := peer.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				logger.Error("Error closing peer connection: %+v", err)
				return
			}
			return

		default:
			if err := t.Decoder.Decode(peer.Conn, &msg); err != nil {
				if errors.Is(err, net.ErrClosed) {
					t.mu.Lock()
					delete(t.peers, peer.ID)
					t.mu.Unlock()

					return
				}
				logger.Error("Error reading message from peer %s: %+v", peer.GetID(), err)
				continue
			}

			t.msgChan.Broadcast(msg)
		}
	}
}

func (t *TCPTransport) Subsribe() broadcaster.Subsriber {
	return t.msgChan.Subsribe()
}

func (t *TCPTransport) Shutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Info("Graceful shutdown started")
	cancel()
	t.mu.Lock()
	for _, peer := range t.peers {
		peer.Close()
	}
	t.mu.Unlock()

	if err := t.listener.Close(); err != nil {
		logger.Error("Error closing listener: %+v", err)
	}

	logger.Info("Closing message channel")
	t.msgChan.Shutdown()
	logger.Info("Graceful shutdown completed")
}

func (t *TCPTransport) Broadcast(msg []byte) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for k, p := range t.peers {
		go func() {
			if err := p.Send(msg); err != nil {
				logger.Error("Error in Broadcast to peer %s", k)
			}
		}()
	}
}

func (t *TCPTransport) Send(peerId string, msg []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	peer, ok := t.peers[peerId]
	if !ok {
		return fmt.Errorf("%s doest exists in peer map\n", peerId)
	}

	if err := peer.Send(msg); err != nil {
		if errors.Is(err, net.ErrClosed) {
			delete(t.peers, peerId)
			return nil
		}
		return err
	}
	return nil
}
