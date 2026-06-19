package clientapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"tsunagi/src/api"
	"tsunagi/src/data"
	"tsunagi/src/database"

	"github.com/minnowo/tsunagi/mod/tcrypto"

	"github.com/flynn/noise"
	"github.com/rs/zerolog/log"
	"github.com/vinovest/sqlx"
)

func (this *ClientApi) apiDeviceInit(w http.ResponseWriter, r *http.Request) {

	var req struct {
		UserID   data.Identifier `json:"user_id"` // TODO: this is from auth token / request authorization
		Password string          `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if len(req.Password) <= 0 {
		api.BadReq(w, "password must not be empty")
		return
	}

	keypair, err := noise.DH25519.GenerateKeypair(rand.Reader)

	if err != nil {
		api.ServerErr(w, "failed to gen user keypair")
		return
	}

	var dev data.Device

	dev.UserID = req.UserID
	dev.PubKey = keypair.Public
	dev.PriKey = keypair.Private

	log.Info().Msg("creating new device")

	encryptedPriKey, err := tcrypto.EncryptWithPassword(dev.PriKey, []byte(req.Password))

	if err != nil {
		api.ServerErr(w, "unable to encrypt private key")
		return
	}

	err = this.DB.Transact(r.Context(), func(ctx context.Context, tx sqlx.Queryable) error {

		var userrow database.UserTable

		if err = this.DB.LoadUserByIdentifier(ctx, tx, req.UserID, &userrow); err != nil {
			return err
		}

		devicerow := database.DeviceTable{
			UserID:          userrow.ID,
			Identifier:      dev.ID,
			PubKey:          dev.PubKey,
			EncryptedPriKey: encryptedPriKey.ToBytes(),
		}

		return this.DB.AddDevice(ctx, tx, &devicerow)
	})

	if err != nil {
		api.ServerErr(w, "error creating a new identity")
	} else {
		// this should not include the private key
		api.WriteJSONObj(w, dev)
	}
}
