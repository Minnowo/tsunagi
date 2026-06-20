package httpapi

import (
	"encoding/json"
	"net/http"
	"tsunagi/src/api"

	"github.com/minnowo/tsunagi/mod/tcrypto"
)

func (h *HttpRelayApi) apiProveChallenge(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Signature []byte `json:"signature"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadReq(w, "invalid json")
		return
	}

	if _, err := tcrypto.ParseAuthToken(req.Signature, h.MacKey[:]); err != nil {
		api.Unauthorized(w)
		return
	}

	api.WriteJSONObj(w, struct {
		Token []byte `json:"token"`
	}{
		Token: req.Signature,
	})
}
