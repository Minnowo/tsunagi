//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/flynn/noise"
)

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

// noiseEncrypt(cipherHandle: number, plaintext: Uint8Array) -> { err, res: Uint8Array }
func noiseEncrypt(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	cs, ok := csStore.get(args[0].Int())
	if !ok {
		return JsErr(ErrInvalidHandle)
	}

	plaintext, ok := JsArrToGo(args[1])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	ciphertext, err := cs.Encrypt(nil, nil, plaintext)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(JsNewUint8Array(ciphertext))
}

// noiseDecrypt(cipherHandle: number, ciphertext: Uint8Array) -> { err, res: Uint8Array }
func noiseDecrypt(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	cs, ok := csStore.get(args[0].Int())
	if !ok {
		return JsErr(ErrInvalidHandle)
	}

	ciphertext, ok := JsArrToGo(args[1])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	plaintext, err := cs.Decrypt(nil, nil, ciphertext)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(JsNewUint8Array(plaintext))
}

// noiseFreeHandshake(handle: number)
func noiseFreeHandshake(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return JsErr(ErrInvalidNumberOfArguments)
	}
	hsStore.delete(args[0].Int())
	return JsRes(nil)
}

// noiseFreeCipher(handle: number)
func noiseFreeCipher(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return JsErr(ErrInvalidNumberOfArguments)
	}
	csStore.delete(args[0].Int())
	return JsRes(nil)
}

func noiseBindings() map[string]any {
	return map[string]any{
		"encrypt":       js.FuncOf(noiseEncrypt),
		"decrypt":       js.FuncOf(noiseDecrypt),
		"freeHandshake": js.FuncOf(noiseFreeHandshake),
		"freeCipher":    js.FuncOf(noiseFreeCipher),
	}
}
