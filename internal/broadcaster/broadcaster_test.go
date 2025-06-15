package broadcaster

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/subhroacharjee/dfst/internal/logger"
)

func TestBroadcaster(t *testing.T) {
	broadcaster := NewBroadcaster(context.Background())

	t.Cleanup(func() {
		broadcaster.Shutdown()
	})

	subsribers := make([]Subsriber, 0)

	subsribers = append(subsribers, broadcaster.Subsribe())
	subsribers = append(subsribers, broadcaster.Subsribe())
	subsribers = append(subsribers, broadcaster.Subsribe())
	subsribers = append(subsribers, broadcaster.Subsribe())

	fmt.Printf("Subsriber length %d \n", len(broadcaster.subsribers))

	messages := make([]Message, 0)
	for range 6 {
		messages = append(messages, Message{
			From:    uuid.New().String(),
			Payload: []byte{0x11, 0x22},
		})
	}

	go func(broadcaster *Broadcaster, messages []Message) {
		time.Sleep(1 * time.Second)

		for idx, msg := range messages {
			logger.Debug("Broadcasting message %d", idx)
			broadcaster.Broadcast(msg)
			logger.Debug("Broadcasting finished %d", idx)

		}
	}(broadcaster, messages)

	for idx, subs := range subsribers {
		t.Run(fmt.Sprintf("Running test for subsriber %d", idx), func(t *testing.T) {
			subId := idx
			subs := subs
			t.Parallel()
			time.Sleep(time.Second)
			for idx := range 6 {
				n, err := rand.Int(rand.Reader, big.NewInt(500))
				if err != nil {
					n = big.NewInt(200)
				}
				logger.Debug("Putting the subsriber %d to sleep for %dms", subId, n.Int64())
				time.Sleep(time.Duration(n.Int64()) * (time.Millisecond))
				logger.Debug("Woke up from sleep %d subsriber", subId)
				msg, ok := <-subs.Consume()
				assert.True(t, ok)
				msg.Acknowledge()
				logger.Debug("Subsriber %d recieved message %d", subId, idx)

				assert.Equal(t, messages[idx].From, msg.From)
			}
		})
	}
}
