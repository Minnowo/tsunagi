package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
	"tsunagi/src/api"

	"github.com/minnowo/tsunagi/mod/tcrypto"
)

func (h *HttpRelayApi) apiGetChallenge(w http.ResponseWriter, r *http.Request) {

	var req struct {
		PubKey           []byte `json:"pub_key"`
		HandshakeInitMsg []byte `json:"handshake_init_msg"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadReq(w, "invalid json")
		return
	}

	state, err := tcrypto.NewResponderHandshakeIN()
	if err != nil {
		api.ServerErr(w, "handshake init failed")
		return
	}

	handshakeDoneMsg, cipher, err := tcrypto.ResponderHandshakeINStep1(req.HandshakeInitMsg, state)
	if err != nil {
		api.BadReq(w, "handshake failed")
		return
	}

	token, err := tcrypto.BuildAuthToken(req.PubKey, time.Hour, h.MacKey[:])
	if err != nil {
		api.BadReq(w, "invalid public key")
		return
	}

	proof, err := cipher.Enc.Encrypt(nil, nil, token)
	if err != nil {
		api.ServerErr(w, "encrypt failed")
		return
	}

	api.WriteJSONObj(w, struct {
		HandshakeDoneMsg []byte `json:"handshake_done_msg"`
		AuthChallenge    []byte `json:"auth_challenge"`
	}{
		HandshakeDoneMsg: handshakeDoneMsg,
		AuthChallenge:    proof,
	})
}
