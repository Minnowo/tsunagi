package httpapi

import (
	"encoding/base64"
	"net/http"
	"strings"
	"tsunagi/src/api"

	"github.com/go-chi/chi/v5"
	"github.com/minnowo/tsunagi/mod/tcrypto"
)

// HttpRelayApi is an HTTP/WebSocket equivalent of the gRPC RelayApi.
// It embeds TsunagiBase so it shares the same backend logic.
type HttpRelayApi struct {
	api.TsunagiBase
}

func (h *HttpRelayApi) getAuthIdentity(r *http.Request) ([]byte, error) {
	auth := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok {
		return nil, tcrypto.ErrMacMismatch
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	return tcrypto.ParseAuthToken(raw, h.MacKey[:])
}

func (h *HttpRelayApi) Register(r chi.Router) {
	r.Post("/auth/challenge", h.apiGetChallenge)
	r.Post("/auth/prove", h.apiProveChallenge)
	r.Get("/ws/client", h.apiConnectClient)
	r.Get("/ws/relay", h.apiConnectRelay)
}
