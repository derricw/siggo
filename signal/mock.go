package signal

import (
	"bufio"
	"bytes"
	"fmt"
	"time"
)

var exampleSendReceipt string = `{"envelope":
  {"source":"+%s",
   "sourceDevice":1,
   "relay":null,
   "timestamp":%d,
   "isReceipt":false,
   "dataMessage":null,
   "syncMessage":
     {"sentMessage":
	   {"timestamp":%d,
	    "message":"%d",
		"expiresInSeconds":0,
		"attachments":[],
		"groupInfo":null,
		"destination":"+%s"
	   },
	 "blockedNumbers":null,
	 "readMessages":null,
	 "type":null
	},
	"callMessage":null,
	"receiptMessage":null
  }
}
`

// MockSignal implements siggo's SignalAPI interface
type MockSignal struct {
	*Signal
	exampleData []byte
	userNumber  string
}

func (ms *MockSignal) Send(dest, msg string) error {
	// send a fake message, just puts in on the "wire"
	timestamp := time.Now().Unix()
	fakeWire := fmt.Sprintf(exampleSendReceipt, ms.userNumber, timestamp, timestamp, msg, dest)
	ms.exampleData = append(ms.exampleData, fakeWire...)
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
	ms.exampleData = []byte{}
	return nil
}

func (ms *MockSignal) ReceiveUntil(done chan struct{}) {
	go func() {
		// better to select with timeout?
		for len(done) == 0 {
			ms.Receive()
			time.Sleep(time.Second * 1)
		}
	}()
}

func NewMockSignal(userNumber string, exampleData []byte) *MockSignal {
	return &MockSignal{
		Signal:      NewSignal(userNumber),
		exampleData: exampleData,
		userNumber:  userNumber,
	}
}
