package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
	"github.com/subhroacharjee/dfst/internal/logger"
)

func (*PeerStore) getValidatedFileInfoAndPath(path string) (string, os.FileInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		logger.Error("Invalid path of file")
		return "", nil, err
	}

	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(absPath); err != nil {
		return "", nil, err
	}

	if fileInfo.IsDir() {
		return "", nil, fmt.Errorf("%s is a directory, currently there is no support for directory upload", absPath)
	}
	return absPath, fileInfo, nil
}

func (s *PeerStore) StoreFile(path string) error {
	absPath, fi, err := s.getValidatedFileInfoAndPath(path)
	if err != nil {
		return err
	}

	fileChecksum, err := CreateFileChecksum(absPath)
	if err != nil {
		return err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var chunkSize int64
	if s.StoreOpts.ChunkSize != 0 {
		chunkSize = int64(s.ChunkSize * 1024 * 1024)
	} else {
		chunkSize = 10 * 1024 * 1024
	}

	chunkMap := make(map[string][]byte)
	noOfChunks := 0
	chunkIds := make([]string, 0)

	fmt.Println(fi.Size() / chunkSize)

	for {

		buffer := make([]byte, chunkSize)
		bytesRead, err := file.Read(buffer[:chunkSize])
		if err != nil && err != io.EOF {
			return err
		}

		if bytesRead == 0 {
			break
		}
		buf := buffer[:bytesRead]
		chunkId := fmt.Sprintf("%s %d", CreateChecksum(buf), noOfChunks)
		chunkMap[chunkId] = buf

		chunkIds = append(chunkIds, chunkId)

		noOfChunks += 1
	}

	// need to resync file checksum and all the chunks so that all nodes are aware
	// of it

	resyncPacketMap := map[string]any{
		"FILEID":    fileChecksum,
		"CHUNKSIZE": chunkSize,
		"NOCHUNKS":  noOfChunks,
		"CHUNKIDS":  strings.Join(chunkIds, ","),
	}

	if err := s.ResyncPeer(RESYNCFILE, resyncPacketMap); err != nil {
		return err
	}
	var wg sync.WaitGroup

	for _, chunkId := range chunkIds {
		chunk := chunkMap[chunkId]

		wg.Add(1)

		go func(chunk []byte, cid string) {
			defer wg.Done()

			subsriber := s.Transport.Subsribe()
			defer subsriber.Unsubscribe()

			// Broadcast the MESSAGE (NTS chunksize)
			if s.BroadcastNTS(cid, len(chunk)) {
				return
			}
			for {
				logger.Debug("Wating for message %s", cid[len(cid)-2:])
				select {
				case msg := <-subsriber.Consume():
					msg.Acknowledge()
					if s.validateMsgAndSendWrite(msg, chunk, cid) {
						return
					}
				case <-time.After(3 * time.Second):
					return
				}
			}
		}(chunk, chunkId)
	}

	wg.Wait()

	return nil
}

func (s *PeerStore) validateMsgAndSendWrite(msg broadcaster.Message, chunk []byte, cid string) bool {
	var recievedMsgPayload map[string]any
	if err := json.Unmarshal(msg.Payload, &recievedMsgPayload); err != nil {
		return false
	}

	opVal, ok := recievedMsgPayload["OPERATION"]
	if !ok {
		return false
	}
	opFloat, ok := opVal.(float64)
	if !ok {
		return false
	}
	op := OPERATION(uint(opFloat)) // Proper conversion to your custom type

	cidVal, ok := recievedMsgPayload["CID"]
	if !ok {
		return false
	}
	rcid, ok := cidVal.(string)
	if !ok {
		return false
	}

	// Perform the logic
	if op != NTS_OK || rcid != cid {
		return false
	}
	logger.Debug("\n\nRECIEVED NTS_OK msgId: %s \n", cid[len(cid)-1:])
	pid := msg.From

	pld := make(map[string]any)
	pld["OPERATION"] = WRITE
	pld["CHUNK"] = chunk
	pld["CID"] = rcid

	pldPacket, err := json.Marshal(pld)
	if err != nil {
		logger.Error("Cant write the packet: %v", err)
		return true
	}
	pldMsg, err := s.Encoder.Encode(broadcaster.Message{
		From:    s.Transport.ID(),
		Payload: pldPacket,
	})
	if err != nil {
		logger.Error("Cant write the packet: %v", err)
		return true
	}

	if err := s.Transport.Send(pid, pldMsg); err != nil {
		logger.Error("%v", err)
		return false
	}

	s.ResyncPeer(RESYNCMEM, map[string]any{
		"CID":    cid,
		"NODEID": msg.From,
	})

	logger.Debug("validateMsgAndSendWrite successful msgId %s", cid[len(cid)-1:])
	return true
}
