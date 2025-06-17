package store

import (
	"encoding/json"
	"fmt"

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

func (s *PeerStore) HandleResyncPeer(operation ResyncOperation, resyncPayload map[string]any) (err error) {
	if operation == RESYNCFILE {
		// update file operation
		// fileId, chunksize, no of chunk, chunkIds store all
		fileId, okFileId := resyncPayload["FILEID"].(string)
		chunksize, okChunkSize := resyncPayload["CHUNKSIZE"].(float64)
		noOfChunks, okNoOfChunks := resyncPayload["NOCHUNKS"].(float64)
		chunkIds, okChunkIds := resyncPayload["CHUNKIDS"].(string)

		if okChunkIds && okChunkSize && okNoOfChunks && okFileId {
			fileSize := noOfChunks * chunksize
			s.KVStore.Add(fmt.Sprintf("%s_FILESIZE", fileId), fmt.Sprintf("%f", fileSize))
			s.KVStore.Add(fmt.Sprintf("%s_NODES", fileId), "")
			s.KVStore.Add(fileId, chunkIds)

		} else {
			return fmt.Errorf("Invalid packet info")
		}

	} else if operation == RESYNCMEM {
		cid, okCid := resyncPayload["CID"].(string)
		nodeId, okNodeId := resyncPayload["NODEID"].(string)

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
