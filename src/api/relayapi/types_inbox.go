package relayapi

import (
	"fmt"
	"sync"
	"tsunagi/src/data"
)

var ErrInboxFull = fmt.Errorf("inbox is full")
var ErrInboxNotFound = fmt.Errorf("inbox not found")

type Box struct {
	Pipe chan []byte
}

type Inbox struct {
	mu    sync.RWMutex
	inbox map[data.Identifier]Box
}

func (i *Inbox) getBox(id data.Identifier) (Box, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	box, ok := i.inbox[id]
	return box, ok
}

func (i *Inbox) getBox2(id data.Identifier) Box {

	box, ok := i.getBox(id)

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

	return box
}

func (i *Inbox) PutMsg(id data.Identifier, msg []byte) error {

	box := i.getBox2(id)

	select {
	case box.Pipe <- msg:
		return nil
	default:
		return ErrInboxFull
	}
}

func (i *Inbox) GetReadPipe(id data.Identifier) (<-chan []byte, error) {

	box := i.getBox2(id)

	return box.Pipe, nil
}
