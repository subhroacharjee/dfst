package p2p

import (
	"net"

	"github.com/subhroacharjee/dfst/internal/logger"
)

// Implementation of another connected node
type TCPPeer struct {
	net.Conn
	ID string

	addr net.Addr

	// if we have dialed and connected to this node then outbound true
	// if we have listening and accepted to this node then outbound false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		ID:       conn.RemoteAddr().String(), // TODO: during handshake transfer the id
		Conn:     conn,
		addr:     conn.RemoteAddr(),
		outbound: outbound,
	}
}

// GetID implements Peer.
func (p *TCPPeer) GetID() string {
	return p.ID
}

// SetID implements Peer.
func (p *TCPPeer) SetID(id string) {
	p.ID = id
}

func (p *TCPPeer) Close() error {
	defer func() {
		logger.Debug("%s peer connection closed", p.addr.String())
	}()
	return p.Conn.Close()
}

func (p *TCPPeer) Send(data []byte) error {
	if _, err := p.Conn.Write(data); err != nil {
		return err
	}
	return nil
}
