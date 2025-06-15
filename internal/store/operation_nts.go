package store

import (
	"encoding/json"
	"syscall"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
	"github.com/subhroacharjee/dfst/internal/logger"
	"github.com/subhroacharjee/dfst/internal/p2p"
)

func (s *PeerStore) BroadcastNTS(cid string, csize int) bool {
	ntsPacketMap := make(map[string]any)
	ntsPacketMap["OPERATION"] = NTS
	ntsPacketMap["CID"] = cid
	ntsPacketMap["CSIZE"] = csize

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

func (s *PeerStore) HandleNTS(peer p2p.Peer, cid string, csize int) error {
	// will use the size to check if the size is okay or not
	// if okay will send NTS_OK back to peer

	hasSpace, err := hasEnoughSpace(s.rootPath, uint64(csize)+1024)
	if err != nil {
		logger.Error("Error occured during space check: %+v", err)
		return err
	}

	if hasSpace {
		// Send nts ok
		payload, err := json.Marshal(map[string]any{
			"OPERATION": NTS_OK,
			"CID":       cid,
		})
		if err != nil {
			return err
		}

		packet, err := s.Encoder.Encode(broadcaster.Message{
			Payload: payload,
		})
		if err != nil {
			return err
		}

		if err := peer.Send(packet); err != nil {
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
