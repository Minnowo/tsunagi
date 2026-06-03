package clientapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/minnowo/log4zero"
)

var logger = log4zero.Get("clientapi")

type ClientApi struct{}

func (this *ClientApi) Register(r chi.Router) {
	r.Get("/ws", this.HandleWebSocket)
}
