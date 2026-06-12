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

func (this *ClientApi) apiIdentityInit(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Password      string `json:"password"`
		NoStorePriKey bool   `json:"no_store_pri_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
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

	var encryptedPriKey []byte

	if !req.NoStorePriKey {

		if len(req.Password) <= 0 {
			api.BadReq(w, "password must not be empty")
			return
		}

		cipherData, err := tcrypto.EncryptWithPassword(acc.PriKey, []byte(req.Password))

		if err != nil {
			api.ServerErr(w, "unable to encrypt private key")
			return
		}

		encryptedPriKey = cipherData.ToBytes()

	} else {

		encryptedPriKey = []byte{}
	}

	err = this.DB.Transact(r.Context(), func(ctx context.Context, tx sqlx.Queryable) error {

		userrow := database.UserTable{
			Identifier:      acc.ID,
			PubKey:          acc.PubKey,
			EncryptedPriKey: encryptedPriKey,
		}

		return this.DB.AddUser(ctx, tx, &userrow)
	})

	if err != nil {
		log.Error().Err(err).Msg("error creating a new identity")
		api.ServerErr(w, "error creating a new identity")
	} else {

		// identity is not device specific, and is owned by the private key.
		// so we provide this to the user who can use it on any device.
		// we optionally store it encrypted with a password.
		var res struct {
			data.User
			PrivateKey []byte `json:"prikey"`
		}
		res.User = acc
		res.PrivateKey = acc.PriKey

		api.WriteJSONObj(w, res)
	}
}
