//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/flynn/noise"
	"github.com/minnowo/tsunagi/mod/tcrypto"
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

// noiseINNewSenderHandshake(pubKey: Uint8Array, priKey: Uint8Array) → { err, res: { handle } }
func noiseINNewSenderHandshake(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	pub, ok := JsArrToGo(args[0])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	pri, ok := JsArrToGo(args[1])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	state, err := tcrypto.NewSenderHandshakeIN(noise.DHKey{Public: pub, Private: pri})
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{"handle": hsStore.put(state)})
}

// noiseINNewResponderHandshake() → { err, res: { handle } }
func noiseINNewResponderHandshake(this js.Value, args []js.Value) any {

	state, err := tcrypto.NewResponderHandshakeIN()
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{"handle": hsStore.put(state)})
}

// noiseINSenderStep1(handle: number) → { err, res: { msg: Uint8Array } }
func noiseINSenderStep1(this js.Value, args []js.Value) any {

	if len(args) < 1 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	state, ok := hsStore.get(args[0].Int())
	if !ok {
		return JsErr(ErrInvalidHandle)
	}

	msg, err := tcrypto.SenderHandshakeINStep1(state)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{"msg": JsNewUint8Array(msg)})
}

// noiseINResponderStep1(msg: Uint8Array, handle: number) → { err, res: { msg: Uint8Array, enc: number, dec: number } }
func noiseINResponderStep1(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	msg, ok := JsArrToGo(args[0])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	state, ok := hsStore.get(args[1].Int())
	if !ok {
		return JsErr(ErrInvalidHandle)
	}

	replyMsg, ciphers, err := tcrypto.ResponderHandshakeINStep1(msg, state)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{
		"msg": JsNewUint8Array(replyMsg),
		"enc": csStore.put(ciphers.Enc),
		"dec": csStore.put(ciphers.Dec),
	})
}

// noiseINSenderStep2(msg: Uint8Array, handle: number) → { err, res: { enc: number, dec: number } }
func noiseINSenderStep2(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	msg, ok := JsArrToGo(args[0])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	state, ok := hsStore.get(args[1].Int())
	if !ok {
		return JsErr(ErrInvalidHandle)
	}

	ciphers, err := tcrypto.SenderHandshakeINStep2(msg, state)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{
		"enc": csStore.put(ciphers.Enc),
		"dec": csStore.put(ciphers.Dec),
	})
}

// noiseEncrypt(cipherHandle: number, plaintext: Uint8Array) → { err, res: Uint8Array }
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

// noiseDecrypt(cipherHandle: number, ciphertext: Uint8Array) → { err, res: Uint8Array }
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

func noiseINBindings() map[string]any {
	return map[string]any{
		"newSenderHandshake":    js.FuncOf(noiseINNewSenderHandshake),
		"newResponderHandshake": js.FuncOf(noiseINNewResponderHandshake),
		"senderStep1":           js.FuncOf(noiseINSenderStep1),
		"responderStep1":        js.FuncOf(noiseINResponderStep1),
		"senderStep2":           js.FuncOf(noiseINSenderStep2),
		"encrypt":               js.FuncOf(noiseEncrypt),
		"decrypt":               js.FuncOf(noiseDecrypt),
		"freeHandshake":         js.FuncOf(noiseFreeHandshake),
		"freeCipher":            js.FuncOf(noiseFreeCipher),
	}
}
