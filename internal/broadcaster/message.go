package broadcaster

type Message struct {
	From    string `json:"from"`
	Payload []byte `json:"payload"`
	AckChan chan any
}

func (m *Message) Acknowledge() {
	m.AckChan <- struct{}{}
}
