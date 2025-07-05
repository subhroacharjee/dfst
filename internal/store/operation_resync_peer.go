package store

import (
	"encoding/json"
	"fmt"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
)

type ResyncPayload struct {
	StorePacket

	ResyncType ResyncOperation `json:"resyncType"`

	// Operation type Sync MEM
	CID    string `json:"cid"`
	NodeId string `json:"nodeId"`

	// Operation type Sync File
	FileID    string `json:"fileId"`
	ChunkSize int64  `json:"chunkSize"`
	NoChunks  int64  `json:"noChunks"`
	ChunkIds  string `json:"chunkIds"`
}

func CreateResyncFilePacket(fileId string, chunkSize int64, noChunks int64, chunkIds string) (ResyncOperation, ResyncPayload) {
	return RESYNCFILE, ResyncPayload{
		StorePacket: StorePacket{
			Operation: RESYNC,
		},
		ResyncType: RESYNCFILE,
		FileID:     fileId,
		ChunkSize:  chunkSize,
		NoChunks:   noChunks,
		ChunkIds:   chunkIds,
	}
}

func CreateResyncMemPacket(cid, nodeId string) (ResyncOperation, ResyncPayload) {
	return RESYNCFILE, ResyncPayload{
		StorePacket: StorePacket{
			Operation: RESYNC,
		},
		ResyncType: RESYNCFILE,
		CID:        cid,
		NodeId:     nodeId,
	}
}

func (s *PeerStore) ResyncPeer(operation ResyncOperation, resyncPayload ResyncPayload) (err error) {
	var message []byte
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

func (s *PeerStore) HandleResyncPeer(resyncPayload ResyncPayload) (err error) {
	if resyncPayload.ResyncType == RESYNCFILE {
		// update file operation
		// fileId, chunksize, no of chunk, chunkIds store all
		fileId, okFileId := resyncPayload.FileID, resyncPayload.FileID != ""
		chunksize, okChunkSize := resyncPayload.ChunkSize, resyncPayload.ChunkSize != 0
		noOfChunks, okNoOfChunks := resyncPayload.NoChunks, resyncPayload.NoChunks != 0
		chunkIds, okChunkIds := resyncPayload.ChunkIds, resyncPayload.ChunkIds != ""

		if okChunkIds && okChunkSize && okNoOfChunks && okFileId {
			fileSize := noOfChunks * chunksize
			s.KVStore.Add(fmt.Sprintf("%s_FILESIZE", fileId), fmt.Sprintf("%d", fileSize))
			s.KVStore.Add(fmt.Sprintf("%s_NODES", fileId), "")
			s.KVStore.Add(fileId, chunkIds)

		} else {
			return fmt.Errorf("Invalid packet info")
		}

	} else if resyncPayload.ResyncType == RESYNCMEM {
		cid, okCid := resyncPayload.CID, resyncPayload.CID != ""
		nodeId, okNodeId := resyncPayload.NodeId, resyncPayload.NodeId != ""

		if okCid && okNodeId {
			fileNodeMetaId := fmt.Sprintf("%s_NODES", cid)
			nodes, ok := s.StoreOpts.KVStore.Get(fileNodeMetaId)
			if ok {
				s.StoreOpts.KVStore.Add(fileNodeMetaId, fmt.Sprintf("%s,%s", nodes, nodeId))
			} else {
				return fmt.Errorf("RESYNCMEM recieved before RESYNCFILE")
			}
		} else {
			return fmt.Errorf("Invalid packet info")
		}
	}

	return nil
}
