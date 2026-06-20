//go:build js && wasm

package main

import "github.com/flynn/noise"

// Concrete handle stores for each type.
// No generics, no sync: wasm is single-threaded.

type hsHandleStore struct {
	next    int
	handles map[int]*noise.HandshakeState
}

func newHsStore() *hsHandleStore {
	return &hsHandleStore{handles: map[int]*noise.HandshakeState{}}
}

func (s *hsHandleStore) put(v *noise.HandshakeState) int {
	id := s.next
	s.next++
	s.handles[id] = v
	return id
}

func (s *hsHandleStore) get(id int) (*noise.HandshakeState, bool) {
	v, ok := s.handles[id]
	return v, ok
}

func (s *hsHandleStore) delete(id int) {
	delete(s.handles, id)
}

type csHandleStore struct {
	next    int
	handles map[int]*noise.CipherState
}

func newCsStore() *csHandleStore {
	return &csHandleStore{handles: map[int]*noise.CipherState{}}
}

func (s *csHandleStore) put(v *noise.CipherState) int {
	id := s.next
	s.next++
	s.handles[id] = v
	return id
}

func (s *csHandleStore) get(id int) (*noise.CipherState, bool) {
	v, ok := s.handles[id]
	return v, ok
}

func (s *csHandleStore) delete(id int) {
	delete(s.handles, id)
}

var (
	hsStore = newHsStore()
	csStore = newCsStore()
)
