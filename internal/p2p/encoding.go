package p2p

import (
	"bytes"
	"io"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
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
	// logger.Debug("Decode is called")

	var b bytes.Buffer

	_, err := io.Copy(&b, r)
	if err != nil {
		return err
	}
	// logger.Debug("No of bytes copied %d", n)

	msg.Payload = b.Bytes()
	// logger.Debug(">>> decode %s", string(msg.Payload))

	return nil
}

func (e DefaultEncoder) Encode(msg broadcaster.Message) ([]byte, error) {
	return msg.Payload, nil
}
