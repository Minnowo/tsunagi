package clientapi

import (
	"tsunagi/src/client"
	"tsunagi/src/database"

	"github.com/go-chi/chi/v5"
	"github.com/minnowo/log4zero"
)

var logger = log4zero.Get("clientapi")

type ClientApi struct {
	client *client.ClientRelayClient
	DB     database.DB
}

func (this *ClientApi) Init() {
	this.client = client.NewClientRelayClient(100)
}

func (this *ClientApi) Register(r chi.Router) {
	r.Get("/ws", this.HandleWebSocket)
	r.Post("/identity/init", this.apiIdentityInit)
	r.Post("/device/init", this.apiDeviceInit)
}
