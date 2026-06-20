package api

import (
	"context"
	"errors"
	"sync"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
)

var (
	ErrClientConnExists = errors.New("client already connected")
)

type ClientConnManager struct {
	connectedUsers map[data.Identifier]*ClientConn
	mu             sync.RWMutex
}

func NewClientConnManager() *ClientConnManager {
	return &ClientConnManager{
		connectedUsers: map[data.Identifier]*ClientConn{},
	}
}

type ClientConn struct {
	SendCh chan *rpc.RelayEvent
	Ctx    context.Context
}

func (this *ClientConnManager) readConn(id data.Identifier) (*ClientConn, bool) {

	this.mu.RLock()
	defer this.mu.RUnlock()
	usr, ok := this.connectedUsers[id]
	return usr, ok
}

func (this *ClientConnManager) AddConn(id data.Identifier, conn *ClientConn) bool {

	this.mu.Lock()
	defer this.mu.Unlock()

	if _, ok := this.connectedUsers[id]; ok {
		return false
	}

	this.connectedUsers[id] = conn

	log.Warn().Str("id", id.String()).Msg("connected user")

	return true
}

func (this *ClientConnManager) RemoveConn(id data.Identifier) {

	this.mu.Lock()
	defer this.mu.Unlock()

	log.Warn().Str("id", id.String()).Msg("disconnected user")
	delete(this.connectedUsers, id)
}

func (this *ClientConnManager) PutRelayMsg(id data.Identifier, msg *rpc.RelayEvent) bool {

	c, ok := this.readConn(id)

	if !ok {
		return false
	}

	select {
	case <-c.Ctx.Done():
		return false

	case c.SendCh <- msg:
		return true
	}
}
