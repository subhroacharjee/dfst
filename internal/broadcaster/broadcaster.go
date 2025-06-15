package broadcaster

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/subhroacharjee/dfst/internal/logger"
)

type Broadcaster struct {
	sync.RWMutex

	subsribers []subscriberImpl
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewBroadcaster(ctx context.Context) *Broadcaster {
	ctx, cancel := context.WithCancel(ctx)
	return &Broadcaster{
		subsribers: make([]subscriberImpl, 0),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (b *Broadcaster) Subsribe() Subsriber {
	b.Lock()
	defer b.Unlock()

	id := uuid.New().String()
	subsriber := subscriberImpl{
		ID:     id,
		Queue:  make(chan Message, 1024),
		Closed: make(chan any),
	}

	b.subsribers = append(b.subsribers, subsriber)
	go func(subsriber subscriberImpl, b *Broadcaster) {
		select {
		case <-subsriber.Closed:
		case <-b.ctx.Done():
		}

		shouldReturn := b.removeSubscriber(subsriber)
		if shouldReturn {
			return
		}
	}(subsriber, b)

	return &subsriber
}

func (b *Broadcaster) removeSubscriber(subsriber subscriberImpl) bool {
	b.Lock()
	defer b.Unlock()
	for idx, subs := range b.subsribers {
		if subs.ID == subsriber.ID {
			close(subs.Queue)
			close(subs.Closed)
			b.subsribers = append(b.subsribers[:idx], b.subsribers[idx+1:]...)
			return true
		}
	}
	return false
}

func (b *Broadcaster) Broadcast(msg Message) {
	b.Lock()
	subsribers := b.subsribers
	b.Unlock()

	var wg sync.WaitGroup

	wg.Add(len(subsribers))
	for _, subsriber := range subsribers {

		ack := make(chan any)

		msg.AckChan = ack

		go func(subsriber subscriberImpl, msg Message) {
			defer func() {
				if r := recover(); r != nil {
					b.removeSubscriber(subsriber)
				}
			}()

			defer wg.Done()

			subsriber.Queue <- msg
			select {
			case <-msg.AckChan:
			case <-time.After(5 * time.Second):
			}
		}(subsriber, msg)
	}
	wg.Wait()
}

func (b *Broadcaster) Shutdown() {
	logger.Debug("Shutdown is called")
	b.cancel()
}
