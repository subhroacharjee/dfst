package broadcaster

type Subsriber interface {
	Consume() <-chan Message
	Unsubscribe()
}

type subscriberImpl struct {
	ID     string
	Queue  chan Message
	Closed chan any
}

// Consume implements Subsriber.
func (s *subscriberImpl) Consume() <-chan Message {
	return s.Queue
}

// Unsubscribe implements Subsriber.
func (s *subscriberImpl) Unsubscribe() {
	s.Closed <- struct{}{}
}
