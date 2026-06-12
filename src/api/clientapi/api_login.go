package clientapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"tsunagi/src/api"
	"tsunagi/src/data"
	"tsunagi/src/database"

	"github.com/minnowo/tsunagi/mod/tcrypto"

	"github.com/rs/zerolog/log"
)

func (this *ClientApi) apiLogin(w http.ResponseWriter, r *http.Request) {

	var req struct {
		DeviceID data.Identifier `json:"user_id"`
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

	var userrow database.UserTable

	if err := this.DB.LoadUserByIdentifier(r.Context(), this.DB.Sqlx(), req.DeviceID, &userrow); err != nil {

		if err == sql.ErrNoRows {
			api.NotFound(w, "the user identity does not exist")
		} else {
			log.Error().Err(err).Msg("error loading user")
			api.ServerErr(w, "error loading user")
		}
		return
	}

	var prikey tcrypto.PasswordEncryptedData

	if err := prikey.FromBytes(userrow.EncryptedPriKey); err != nil {
		log.Error().Err(err).Msg("error parsing encrypted data")
		api.ServerErr(w, "error loading user")
		return
	}

	privatekey, err := tcrypto.DecryptWithPassword(&prikey, []byte(req.Password))

	if err != nil {
		api.BadReq(w, "unable to decrypt")
		return
	}

	var user data.User

	user.ID = userrow.Identifier
	user.PubKey = userrow.PubKey
	user.PriKey = privatekey

	// TODO: finish this
	// This should be loading a device probably not the user identity.
	// The user identity is only to be used for device revocation / identity stuff?

}
