package p2p

import (
	"bufio"
	"io"
	"net"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
	"github.com/subhroacharjee/dfst/internal/logger"
)

type Decoder interface {
	Decode(io.Reader, *broadcaster.Message) error
}

type Encoder interface {
	Encode(broadcaster.Message) ([]byte, error)
}

type (
	DefaultDecoder struct{}
	DefaultEncoder struct{}
)

func (d DefaultDecoder) Decode(r io.Reader, msg *broadcaster.Message) error {
	logger.Debug("Decode is called")

	// var b bytes.Buffer

	// This is blocking the whole thing need to handle it
	// n, err := io.Copy(&b, r)
	// if err != nil {
	// 	return err
	// }
	// if n == 0 {
	// 	return net.ErrClosed
	// }
	// logger.Debug("No of bytes copied %d", n)
	//
	// msg.Payload = b.Bytes()
	// logger.Debug(">>> decode %s", string(msg.Payload))

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return scanner.Err()
		} else {
			return net.ErrClosed
		}
	}

	msg.Payload = scanner.Bytes()
	logger.Debug("%v", msg)

	return nil
}

func (e DefaultEncoder) Encode(msg broadcaster.Message) ([]byte, error) {
	return msg.Payload, nil
}
