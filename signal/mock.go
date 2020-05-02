package signal

import (
	"bufio"
	"bytes"
)

// MockSignal implements siggo's SignalAPI interface
type MockSignal struct {
	*Signal
	exampleData []byte
}

func (ms *MockSignal) Send(dest, msg string) error {
	// send a fake message
	return nil
}

func (ms *MockSignal) Receive() error {
	r := bytes.NewReader(ms.exampleData)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		wire := scanner.Bytes()
		err := ms.ProcessWire(wire)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewMockSignal(exampleData []byte) *MockSignal {
	return &MockSignal{
		Signal:      NewSignal(""),
		exampleData: exampleData,
	}
}
