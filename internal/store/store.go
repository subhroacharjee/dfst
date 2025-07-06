package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/subhroacharjee/dfst/internal/broadcaster"
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

	StorePacket struct {
		Operation OPERATION `json:"operation"`
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

func (p *PeerStore) setupTransport(ctx context.Context) error {
	return p.Transport.ListenAndAccept(ctx)
}

func (p *PeerStore) Start(ctx context.Context) error {
	if err := p.setupTransport(ctx); err != nil {
		return err
	}

	p.ConsumeAndOperate(ctx)
	return nil
}

func (p *PeerStore) ConsumeAndOperate(ctx context.Context) {
	sub := p.Transport.Subsribe()
	for {
		select {
		case <-ctx.Done():
			// call shutdown on message queue.
			sub.Unsubscribe()
		case msg := <-sub.Consume():
			logger.Debug("New message recieved")
			go p.handleMessage(&msg)
		}
	}
}

func (p *PeerStore) handleMessage(msg *broadcaster.Message) {
	msg.Acknowledge()
	var payload map[string]any

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		logger.Error("error Unmarshalling the payload: %v", err)
		return
	}

	operation, exists := payload["operation"]
	if !exists {
		logger.Error("Invalid message packet found")
		return
	}

	operation, ok := operation.(float64)
	if !ok {
		logger.Error("Invalid operation type")
		return
	}
	op, ok := operation.(OPERATION)
	if !ok {
		logger.Error("Invalid operation type")
		return
	}

	if op == RESYNC {
		var resyncPayload ResyncPayload
		if err := json.Unmarshal(msg.Payload, &resyncPayload); err != nil {
			logger.Error("Invalid resync payload, Unmarshal error: %v", err)
			return
		}

		if err := p.HandleResyncPeer(resyncPayload); err != nil {
			logger.Error("Error handle resync peer: %v", err)
			return
		}

	} else if op == NTS {
		var ntsPayload NTSPayload
		if err := json.Unmarshal(msg.Payload, &ntsPayload); err != nil {
			logger.Error("Invalid nts payload, unmarshal error: %v", err)
			return
		}
		p.HandleNTS(func(packet []byte) error {
			return p.Transport.Send(msg.From, packet)
		}, ntsPayload)
	} else if op == WRITE {
		var writePayload WritePayload
		if err := json.Unmarshal(msg.Payload, &writePayload); err != nil {
			logger.Error("Invalid write payload, unmarshal error: %v", err)
			return
		}
		if err := p.Write(writePayload); err != nil {
			// TODO: add support to propagate this error to all peers
			logger.Error("Error in writing chunk: %v", err)
			return
		}
	}
}
