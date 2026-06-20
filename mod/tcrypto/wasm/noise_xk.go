//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/flynn/noise"
	"github.com/minnowo/tsunagi/mod/tcrypto"
)

// noiseXKNewSenderHandshake(responderPubKey: Uint8Array, pubKey: Uint8Array, priKey: Uint8Array) -> { err, res: { handle } }
func noiseXKNewSenderHandshake(this js.Value, args []js.Value) any {

	if len(args) < 3 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	responderPub, ok := JsArrToGo(args[0])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	pub, ok := JsArrToGo(args[1])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	pri, ok := JsArrToGo(args[2])
	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	state, err := tcrypto.NewSenderHandshakeXK(responderPub, noise.DHKey{Public: pub, Private: pri})
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{"handle": hsStore.put(state)})
}

// noiseXKNewResponderHandshake(pubKey: Uint8Array, priKey: Uint8Array) -> { err, res: { handle } }
func noiseXKNewResponderHandshake(this js.Value, args []js.Value) any {

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

	state, err := tcrypto.NewResponderHandshakeXK(noise.DHKey{Public: pub, Private: pri})
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{"handle": hsStore.put(state)})
}

// noiseXKSenderStep1(handle: number) -> { err, res: { msg: Uint8Array } }
func noiseXKSenderStep1(this js.Value, args []js.Value) any {

	if len(args) < 1 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	state, ok := hsStore.get(args[0].Int())
	if !ok {
		return JsErr(ErrInvalidHandle)
	}

	msg, err := tcrypto.SenderHandshakeXKStep1(state)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{"msg": JsNewUint8Array(msg)})
}

// noiseXKResponderStep1(msg: Uint8Array, handle: number) -> { err, res: { msg: Uint8Array } }
func noiseXKResponderStep1(this js.Value, args []js.Value) any {

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

	replyMsg, err := tcrypto.ResponderHandshakeXKStep1(msg, state)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{"msg": JsNewUint8Array(replyMsg)})
}

// noiseXKSenderStep2(msg: Uint8Array, handle: number) -> { err, res: nil }
func noiseXKSenderStep2(this js.Value, args []js.Value) any {

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

	if err := tcrypto.SenderHandshakeXKStep2(msg, state); err != nil {
		return JsErr(err)
	}

	return JsRes(nil)
}

// noiseXKResponderStep2(msg: Uint8Array, handle: number) -> { err, res: { enc: number, dec: number } }
func noiseXKResponderStep2(this js.Value, args []js.Value) any {

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

	ciphers, err := tcrypto.ResponderHandshakeXKStep2(msg, state)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{
		"enc": csStore.put(ciphers.Enc),
		"dec": csStore.put(ciphers.Dec),
	})
}

// noiseXKSenderStep3(handle: number) -> { err, res: { msg: Uint8Array, enc: number, dec: number } }
func noiseXKSenderStep3(this js.Value, args []js.Value) any {

	if len(args) < 1 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	state, ok := hsStore.get(args[0].Int())
	if !ok {
		return JsErr(ErrInvalidHandle)
	}

	msg, ciphers, err := tcrypto.SenderHandshakeXKStep3(state)
	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{
		"msg": JsNewUint8Array(msg),
		"enc": csStore.put(ciphers.Enc),
		"dec": csStore.put(ciphers.Dec),
	})
}

func noiseXKBindings() map[string]any {
	return map[string]any{
		"newSenderHandshake":    js.FuncOf(noiseXKNewSenderHandshake),
		"newResponderHandshake": js.FuncOf(noiseXKNewResponderHandshake),
		"senderStep1":           js.FuncOf(noiseXKSenderStep1),
		"responderStep1":        js.FuncOf(noiseXKResponderStep1),
		"senderStep2":           js.FuncOf(noiseXKSenderStep2),
		"responderStep2":        js.FuncOf(noiseXKResponderStep2),
		"senderStep3":           js.FuncOf(noiseXKSenderStep3),
	}
}
