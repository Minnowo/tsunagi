//go:build js && wasm

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"syscall/js"

	"github.com/flynn/noise"
	"github.com/minnowo/tsunagi/mod/tcrypto"
)

var (
	ErrInvalidNumberOfArguments = fmt.Errorf("invalid number of arguments")
	ErrArgumentWasntUint8Array  = fmt.Errorf("expected Uint8Array or Uint8ClampedArray")
	ErrInvalidHandle            = fmt.Errorf("invalid handle")
)

var (
	b64 = base64.StdEncoding
)

func JsErr(err error) map[string]any {
	return map[string]any{
		"err": err.Error(),
		"res": nil,
	}
}

func JsRes(res any) map[string]any {
	return map[string]any{
		"err": nil,
		"res": res,
	}
}

func JsNewUint8Array(bytes []byte) js.Value {

	uint8Array := js.Global().Get("Uint8Array").New(len(bytes))
	js.CopyBytesToJS(uint8Array, bytes)

	return uint8Array
}

func JsArrToGo(val js.Value) ([]byte, bool) {

	name := val.Get("constructor").Get("name").String()

	if name != "Uint8Array" && name != "Uint8ClampedArray" {
		return nil, false
	}

	dst := make([]byte, val.Length())
	js.CopyBytesToGo(dst, val)

	return dst, true
}

func decryptWithPassword(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	dataBytes := args[0]
	passStr := args[1].String()

	raw, ok := JsArrToGo(dataBytes)

	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	var encrypted tcrypto.PasswordEncryptedData

	if err := encrypted.FromBytes(raw); err != nil {
		return JsErr(err)
	}

	decrypted, err := tcrypto.DecryptWithPassword(&encrypted, []byte(passStr))

	if err != nil {
		return JsErr(err)
	}

	return JsRes(JsNewUint8Array(decrypted))
}

func encryptStrWithPassword(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	dataStr := args[0].String()
	passStr := args[1].String()

	return _encryptWithPassword([]byte(dataStr), []byte(passStr))
}

func encryptWithPassword(this js.Value, args []js.Value) any {

	if len(args) < 2 {
		return JsErr(ErrInvalidNumberOfArguments)
	}

	dataBytes := args[0]
	passStr := args[1].String()

	raw, ok := JsArrToGo(dataBytes)

	if !ok {
		return JsErr(ErrArgumentWasntUint8Array)
	}

	return _encryptWithPassword(raw, []byte(passStr))
}

func _encryptWithPassword(raw, pass []byte) any {

	encrypted, err := tcrypto.EncryptWithPassword(raw, pass)

	if err != nil {
		return JsErr(err)
	}

	bytes := encrypted.ToBytes()

	return JsRes(JsNewUint8Array(bytes))
}

func generateKeypair(this js.Value, args []js.Value) any {

	keypair, err := noise.DH25519.GenerateKeypair(rand.Reader)

	if err != nil {
		return JsErr(err)
	}

	return JsRes(map[string]any{
		"public":  JsNewUint8Array(keypair.Public),
		"private": JsNewUint8Array(keypair.Private),
	})
}

func main() {

	js.Global().Set("tcrypto", map[string]any{
		"generateKeypair":        js.FuncOf(generateKeypair),
		"encryptWithPassword":    js.FuncOf(encryptWithPassword),
		"encryptStrWithPassword": js.FuncOf(encryptStrWithPassword),
		"decryptWithPassword":    js.FuncOf(decryptWithPassword),
		"noiseIN":                noiseINBindings(),
		"noiseXK":                noiseXKBindings(),
		"noise":                  noiseBindings(),
	})

	select {}
}
