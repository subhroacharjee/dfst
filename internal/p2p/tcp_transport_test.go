package p2p

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TCPTransport(t *testing.T) {
	trOpts := TCPTransportOpts{
		ListenerAddr:  ":4000",
		Decoder:       DefaultDecoder{},
		HandShakeFunc: DefaultHandShake,
	}

	tr := NewTCPTransport(trOpts)
	ctx, cancel := context.WithCancel(context.Background())
	err := tr.ListenAndAccept(ctx)
	fmt.Println(err)
	assert.Nil(t, err)
	tr.Shutdown(cancel)
}
