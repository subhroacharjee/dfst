package store

import (
	"encoding/json"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
)

func (s *PeerStore) ResyncPeer(operation ResyncOperation, resyncPayload map[string]any) (err error) {
	var message []byte
	resyncPayload["OPERATION"] = RESYNC
	resyncPayload["RESYNCTYPE"] = operation

	payload, err := json.Marshal(resyncPayload)
	message, err = s.Encoder.Encode(broadcaster.Message{
		From:    s.StoreOpts.Transport.ID(),
		Payload: payload,
	})
	if err != nil {
		return err
	}

	s.StoreOpts.Transport.Broadcast(message)

	return nil
}
