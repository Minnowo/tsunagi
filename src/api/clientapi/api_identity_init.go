package clientapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"tsunagi/src/api"
	"tsunagi/src/data"
	"tsunagi/src/database"
	"tsunagi/src/tcrypto"

	"github.com/flynn/noise"
	"github.com/rs/zerolog/log"
	"github.com/vinovest/sqlx"
)

func (this *ClientApi) apiIdentityInit(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Password string `json:"password"`
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

	var acc data.User

	acc.ID.GenNew()
	acc.PubKey = keypair.Public
	acc.PriKey = keypair.Private

	log.Info().Msg("creating new identity")

	encryptedPriKey, err := tcrypto.EncryptWithPassword(acc.PriKey, []byte(req.Password))

	if err != nil {
		api.ServerErr(w, "unable to encrypt private key")
		return
	}

	err = this.DB.Transact(r.Context(), func(ctx context.Context, tx sqlx.Queryable) error {

		userrow := database.UserTable{
			Identifier:      acc.ID,
			PubKey:          acc.PubKey,
			EncryptedPriKey: encryptedPriKey.ToBytes(),
		}

		return this.DB.AddUser(ctx, tx, &userrow)
	})

	if err != nil {
		log.Error().Err(err).Msg("error creating a new identity")
		api.ServerErr(w, "error creating a new identity")
	} else {
		// this should not include the private key
		api.WriteJSONObj(w, acc)
	}
}
