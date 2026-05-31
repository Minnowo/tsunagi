package relayapi

import (
	"fmt"
	"sync"
	"tsunagi/src/data"
)

var ErrInboxFull = fmt.Errorf("inbox is full")
var ErrInboxNotFound = fmt.Errorf("inbox not found")

type Box struct{
	Pipe chan []byte
}

type Inbox struct {
	mu sync.RWMutex
	inbox map[data.Identifier]Box
}

func (i *Inbox) getBox(id data.Identifier) (Box, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	box , ok := i.inbox[id] 
	return	box , ok
}

func(i *Inbox) PutMsg(id data.Identifier, msg []byte) error {

	box , ok := i.getBox(id)

	if !ok {

		box = Box{
			Pipe: make(chan []byte, 100),
		}

		func() {
			i.mu.Lock()
			defer i.mu.Unlock()

			i.inbox[id] = box
		}()
	}

	select {
	case box.Pipe <- msg:
		return nil
	default:
		return ErrInboxFull
	}
}

func(i *Inbox) GetReadPipe(id data.Identifier) ( <-chan []byte, error) {

	box, ok := i.getBox(id)

	if !ok {
		return nil, ErrInboxNotFound
	}

	return box.Pipe, nil
}
