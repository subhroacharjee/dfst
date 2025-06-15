package store

import (
	"os"
	"path/filepath"

	"github.com/subhroacharjee/dfst/internal/kvstore"
	"github.com/subhroacharjee/dfst/internal/logger"
	"github.com/subhroacharjee/dfst/internal/p2p"
)

type (
	StoreOpts struct {
		ChunkSize uint // Chunksize in MB
		Transport p2p.Transport
		Encoder   p2p.Encoder
		KVStore   kvstore.KVStore
	}

	PeerStore struct {
		StoreOpts
		rootPath string
	}
)

func NewStore(opts StoreOpts) *PeerStore {
	var rootPath string
	err := os.MkdirAll("./data", 0755)
	if err != nil {
		logger.Error("Error in creating data directory %+v.\n using /tmp", err)
		rootPath = "/tmp"
	}
	rootPath, err = filepath.Abs("./data")
	if err != nil {
		logger.Error("Error in creating abs path of ./data %+v.\n using /tmp", err)
		rootPath = "/tmp"
	}
	return &PeerStore{
		StoreOpts: opts,
		rootPath:  rootPath,
	}
}

func (p *PeerStore) SetPath(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	rootPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	p.rootPath = rootPath
	return nil
}
