package store

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/subhroacharjee/dfst/internal/broadcaster"
	"github.com/subhroacharjee/dfst/internal/logger"
	"github.com/subhroacharjee/dfst/internal/p2p"
)

const (
	CHUNK_SIZE   uint   = 1
	NO_OF_CHUNKS uint64 = 10
)

type MockTransport struct {
	id string

	messageChan *broadcaster.Broadcaster

	broadcastedMsgs []broadcaster.Message
	sentMsgs        map[string]map[string]any
	Decoder         p2p.Decoder
	mu              sync.RWMutex
}

func createTestFile(filename string, size uint64) (string, error) {
	err := os.MkdirAll("./tmp", 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create ./tmp directory: %w", err)
	}

	// Create full file path
	path := filepath.Join("./tmp", filename)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	const bufferSize = 1024
	buf := make([]byte, bufferSize)

	var written uint64
	for written < size {
		toWrite := bufferSize
		if size-written < bufferSize {
			toWrite = int(size - written)
		}

		_, err := rand.Read(buf[:toWrite])
		if err != nil {
			return "", fmt.Errorf("failed to generate random data: %w", err)
		}

		n, err := file.Write(buf[:toWrite])
		if err != nil {
			return "", fmt.Errorf("failed to write to file: %w", err)
		}

		written += uint64(n)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

func TestStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mockTransport := MockTransport{
		id:              "mock-transport",
		messageChan:     broadcaster.NewBroadcaster(ctx),
		Decoder:         p2p.DefaultDecoder{},
		broadcastedMsgs: make([]broadcaster.Message, 0),
		sentMsgs:        make(map[string]map[string]any),
	}

	opts := StoreOpts{
		ChunkSize: CHUNK_SIZE,
		Transport: &mockTransport,
		Encoder:   p2p.DefaultEncoder{},
	}

	store := NewStore(opts)

	assert.NotNil(t, store)

	testPath, err := createTestFile("test.txt", NO_OF_CHUNKS*1024*1024)
	if err != nil {
		t.Errorf("%v : %v", reflect.TypeOf(err), err)
	}

	t.Cleanup(func() {
		mockTransport.Shutdown(cancel)
		os.RemoveAll("./tmp")
		os.RemoveAll(store.rootPath)
	})

	go mockTransport.sendNTSOkReply(NO_OF_CHUNKS)

	if err := store.StoreFile(testPath); err != nil {
		fmt.Printf("%v: %v\n", reflect.TypeOf(err), err)
		assert.NotNil(t, err)

	}

	assert.Equal(t, int(NO_OF_CHUNKS)*2+1, len(mockTransport.broadcastedMsgs))
	assert.Equal(t, int(NO_OF_CHUNKS), len(mockTransport.sentMsgs))

	n, err := rand.Int(rand.Reader, big.NewInt(int64(NO_OF_CHUNKS)))
	if err != nil {
		t.Error(err)
		return
	}

	counter := 0
	var msg map[string]any

	for _, val := range mockTransport.sentMsgs {
		counter++
		if counter == int(n.Int64()) {
			msg = val
			break
		}
	}

	bt, err := json.Marshal(msg)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	var wtpld WritePayload

	if err := json.Unmarshal(bt, &wtpld); err != nil {
		t.Error(err)
		t.Fail()
	}

	if err := store.Write(wtpld); err != nil {
		t.Errorf("Error occured during writing chunk: %v", err)
		t.Fail()
	}

	absPath, err := filepath.Abs(filepath.Join(store.rootPath, wtpld.CID, fmt.Sprintf("%s.cnk", wtpld.CID)))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if _, err := os.Stat(absPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Errorf("The required file doesnt exists")
			t.Fail()
		} else {
			t.Errorf("Some other error %v", err)
			t.Fail()
		}
	}
}

func (m *MockTransport) sendNTSOkReply(noOfChunks uint64) {
	for {
		time.Sleep(2 * time.Second)
		m.mu.Lock()
		if len(m.broadcastedMsgs) > int(noOfChunks) {
			break
		}
		m.mu.Unlock()

	}

	broadcastedMsgs := m.broadcastedMsgs

	m.mu.Unlock()
	var wg sync.WaitGroup

	for idx, msg := range broadcastedMsgs {

		wg.Add(1)
		go func(mockPeerId int, msg broadcaster.Message) {
			defer wg.Done()
			n, err := rand.Int(rand.Reader, big.NewInt(100))
			if err != nil {
				n = big.NewInt(50)
			}
			time.Sleep(time.Duration(n.Int64()) * time.Millisecond)

			var originalMap map[string]any
			err = json.Unmarshal(msg.Payload, &originalMap)
			if err != nil {
				fmt.Println("error", err)
				return
			}
			val, exists := originalMap["operation"]
			if !exists {
				return
			}
			operation, ok := val.(float64)

			if ok && OPERATION(operation) == NTS {

				cid, _ := originalMap["cid"].(string)

				replyMessageInBytes, err := json.Marshal(MakeNTSOkPacket(cid))
				if err != nil {
					fmt.Println("error", err)
					panic(err)
				}

				reply := broadcaster.Message{
					From:    fmt.Sprintf("msg_%d", mockPeerId),
					Payload: replyMessageInBytes,
				}

				m.messageChan.Broadcast(reply)
			}
		}(idx, msg)

	}

	wg.Wait()
}

// Broadcast implements broadcaster.Transport.
func (m *MockTransport) Broadcast(b []byte) {
	var msg broadcaster.Message
	if err := m.Decoder.Decode(bytes.NewReader(b), &msg); err != nil {
		panic(err)
	}
	logger.Debug("Calling broadcast messages")

	m.mu.Lock()
	defer m.mu.Unlock()

	m.broadcastedMsgs = append(m.broadcastedMsgs, msg)
}

// Consume implements broadcaster.Transport.
func (m *MockTransport) Subsribe() broadcaster.Subsriber {
	return m.messageChan.Subsribe()
}

// Dial implements broadcaster.Transport.
func (m *MockTransport) Dial(context.Context, string) error {
	panic("unimplemented")
}

// ID implements broadcaster.Transport.
func (m *MockTransport) ID() string {
	return m.id
}

// ListenAndAccept implements broadcaster.Transport.
func (m *MockTransport) ListenAndAccept(context.Context) error {
	panic("unimplemented")
}

// Send implements broadcaster.Transport.
func (m *MockTransport) Send(peerID string, msg []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var originalMapPacket map[string]any

	if err := json.Unmarshal(msg, &originalMapPacket); err != nil {
		return err
	}
	m.sentMsgs[peerID] = originalMapPacket
	return nil
}

// Shutdown implements broadcaster.Transport.
func (m *MockTransport) Shutdown(cancel context.CancelFunc) {
	m.messageChan.Shutdown()
	cancel()
}
