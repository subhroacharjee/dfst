package store

import (
	"encoding/json"
	"syscall"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
	"github.com/subhroacharjee/dfst/internal/logger"
)

type (
	NTSPayload struct {
		StorePacket

		CID   string `json:"cid"`
		CSize int    `json:"csize"`
	}

	NTSOkPayload struct {
		StorePacket
		CID string `json:"cid"`
	}
)

func MakeNTSPacket(cid string, csize int) NTSPayload {
	return NTSPayload{
		StorePacket: StorePacket{
			Operation: NTS,
		},
		CID:   cid,
		CSize: csize,
	}
}

func MakeNTSOkPacket(cid string) NTSOkPayload {
	return NTSOkPayload{
		StorePacket: StorePacket{
			Operation: NTS_OK,
		},
		CID: cid,
	}
}

func (s *PeerStore) BroadcastNTS(ntsPacketMap NTSPayload) bool {
	ntsPacket, err := json.Marshal(ntsPacketMap)
	if err != nil {
		logger.Error("Cant broadcast nts packet map: %v", err)
		return true
	}
	ntsMsg, err := s.Encoder.Encode(broadcaster.Message{
		From:    s.Transport.ID(),
		Payload: ntsPacket,
	})
	if err != nil {
		logger.Error("Cant broadcast nts packet map: %v", err)

		return true
	}
	s.StoreOpts.Transport.Broadcast(ntsMsg)
	return false
}

func (s *PeerStore) HandleNTS(send func([]byte) error, payload NTSPayload) error {
	// will use the size to check if the size is okay or not
	// if okay will send NTS_OK back to peer

	cid, csize := payload.CID, payload.CSize
	hasSpace, err := hasEnoughSpace(s.rootPath, uint64(csize)+1024)
	if err != nil {
		logger.Error("Error occured during space check: %+v", err)
		return err
	}

	if hasSpace {
		// Send nts ok
		payload, err := json.Marshal(MakeNTSOkPacket(cid))
		if err != nil {
			return err
		}

		packet, err := s.Encoder.Encode(broadcaster.Message{
			Payload: payload,
		})
		if err != nil {
			return err
		}

		if err := send(packet); err != nil {
			return err
		}
	}
	return nil
}

func hasEnoughSpace(path string, requiredBytes uint64) (bool, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return false, err
	}

	// Available blocks * size per block = available space in bytes
	available := stat.Bavail * uint64(stat.Bsize)

	return available >= requiredBytes, nil
}
