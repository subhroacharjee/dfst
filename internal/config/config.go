package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/subhroacharjee/dfst/internal/kvstore"
	"github.com/subhroacharjee/dfst/internal/p2p"
	"github.com/subhroacharjee/dfst/internal/store"
)

type (
	TransportType string
	HandShakeAlgo string
	EncoderAlgo   string
	KVStoreType   string
)

const (
	// TransportType
	EMPTY_TRANSPORT_TYPE TransportType = ""
	TCP                  TransportType = "TCP"

	// HandShakeAlgo
	EMPTY_HANDSHAKE   HandShakeAlgo = ""
	DEFAULT_HANDSHAKE HandShakeAlgo = "DEFAULT_HANDSHAKE"

	// EncoderAlgo
	EMPTY_ENCODER   EncoderAlgo = ""
	DEFAULT_ENCODER EncoderAlgo = "DEFAULT_ENCODER"

	// KVStoreType
	EMPTY_KVSTORE   KVStoreType = ""
	DEFAULT_KVSTORE KVStoreType = "DEFAULT_KVSTORE"
)

type Config struct {
	TransportType TransportType `json:"transportType"`
	HandShakeAlgo HandShakeAlgo `json:"handshakeAlgo"`
	EncoderAlgo   EncoderAlgo   `json:"encoderAlg"`
	KVStoreType   KVStoreType   `json:"kvStoreType"`

	ListenAddr string `json:"listenAddr"`
	PeerId     string `json:"peerId"`
	ChunkSize  uint   `json:"chunkSize"`

	PeerAddrs []string `json:"peerAddrs"`

	store     kvstore.KVStore
	decoder   p2p.Decoder
	encoder   p2p.Encoder
	handshake p2p.HandShakeFunc

	transport p2p.Transport
}

func GetDefaultConfig() *Config {
	c := Config{
		TransportType: TCP,
		HandShakeAlgo: DEFAULT_HANDSHAKE,
		EncoderAlgo:   DEFAULT_ENCODER,
		KVStoreType:   DEFAULT_KVSTORE,

		ListenAddr: ":4000",
		PeerId:     uuid.NewString(),
		ChunkSize:  10,

		PeerAddrs: make([]string, 0),

		store:     kvstore.NewInMemoryStore(),
		decoder:   p2p.DefaultDecoder{},
		encoder:   p2p.DefaultEncoder{},
		handshake: p2p.DefaultHandShake,
	}
	c.transport = p2p.NewTCPTransport(*c.GetTransportOpts())

	return &c
}

func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var config Config

	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	if config.HandShakeAlgo == DEFAULT_HANDSHAKE || config.HandShakeAlgo == EMPTY_HANDSHAKE {
		config.handshake = p2p.DefaultHandShake
	} else {
		return nil, fmt.Errorf("invalid handshake algorithm")
	}

	if config.EncoderAlgo == DEFAULT_ENCODER || config.EncoderAlgo == EMPTY_ENCODER {
		config.decoder = p2p.DefaultDecoder{}
		config.encoder = p2p.DefaultEncoder{}
	} else {
		return nil, fmt.Errorf("invalid encoder algorithm")
	}

	if config.KVStoreType == DEFAULT_KVSTORE || config.KVStoreType == EMPTY_KVSTORE {
		config.store = kvstore.NewInMemoryStore()
	} else {
		return nil, fmt.Errorf("invalid kvstore type")
	}

	if config.TransportType == TCP || config.TransportType == EMPTY_TRANSPORT_TYPE {
		config.transport = p2p.NewTCPTransport(*config.GetTransportOpts())
	} else {
		return nil, fmt.Errorf("invalid transport type")
	}

	return &config, nil
}

func (c Config) GetTransportOpts() *p2p.TransportOpts {
	return &p2p.TransportOpts{
		PeerID:        c.PeerId,
		ListenerAddr:  c.ListenAddr,
		Decoder:       c.decoder,
		HandShakeFunc: c.handshake,
	}
}

func (c Config) GetStoreOpts() *store.StoreOpts {
	return &store.StoreOpts{
		ChunkSize: c.ChunkSize,
		Encoder:   c.encoder,
		KVStore:   c.store,
		Transport: c.transport,
	}
}
