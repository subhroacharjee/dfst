package main

import (
	"context"
	"fmt"
	"log"

	"github.com/subhroacharjee/dfst/internal/p2p"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	trOpts := p2p.TCPTransportOpts{
		ListenerAddr:  ":4000",
		Decoder:       p2p.DefaultDecoder{},
		HandShakeFunc: p2p.DefaultHandShake,
	}

	tr := p2p.NewTCPTransport(trOpts)

	err = tr.ListenAndAccept(context.Background())
	return err
}
