package relayapi

// import (
// 	"net/http"
// 	"time"
// 	"tsunagi/src/api"

// 	"github.com/go-chi/chi/v5"
// )


// func(this*RelayApi) apiInbox(w http.ResponseWriter, r *http.Request) {

// 	var deviceID data.Identifier

// 	if err := deviceID.FromString(chi.URLParam(r, "deviceID")); err!=nil{
// 		api.BadReq(w, "invalid identifier")
// 		return
// 	}

// 	pipe, err := this.inbox.GetReadPipe(deviceID)

// 	if err != nil {
// 		api.NotFound(w, "the inbox was not found")
// 		return
// 	}

// 	rc := http.NewResponseController(w)
// 	rc.SetWriteDeadline(time.Time{})

// 	select {
// 	case msg := <-pipe:
// 		w.Header().Set("Content-Type", "application/octet-stream")
// 		w.Write(msg)

// 	case <-r.Context().Done():
// 		// client disconnected
// 		return
// 	}
// }
