package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/subhroacharjee/dfst/internal/logger"
)

type WritePayload struct {
	StorePacket

	Chunk    []byte `json:"chunk"`
	CID      string `json:"cid"`
	Checksum string `json:"checksum,omitempty"`
}

func NewWritePayload(chunk []byte, cid string) WritePayload {
	packet := WritePayload{
		StorePacket: StorePacket{
			Operation: WRITE,
		},

		Chunk: chunk,
		CID:   cid,
	}

	bt, _ := json.Marshal(packet)
	packet.Checksum = CreateChecksum(bt)

	return packet
}

func (w WritePayload) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

func (w WritePayload) ValidateChecksum() error {
	temp := WritePayload{
		StorePacket: w.StorePacket,
		Chunk:       w.Chunk,
		CID:         w.CID,
	}

	bt, err := temp.Marshal()
	if err != nil {
		return err
	}

	checksum := CreateChecksum(bt)
	if checksum != w.Checksum {
		return fmt.Errorf("Invalid Checksum")
	}
	return nil
}

func (s *PeerStore) Write(payload WritePayload) error {
	if err := payload.ValidateChecksum(); err != nil {
		return err
	}
	logger.Info("Checksum validation successful")

	chunkDirFullPath, err := filepath.Abs(filepath.Join(s.rootPath, payload.CID))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(chunkDirFullPath, 0755); err != nil {
		return err
	}

	chunkPath := filepath.Join(chunkDirFullPath, fmt.Sprintf("%s.cnk", payload.CID))

	if err := os.WriteFile(chunkPath, payload.Chunk, 0644); err != nil {
		return err
	}

	logger.Info("Writing chunk successful")

	return nil
}
