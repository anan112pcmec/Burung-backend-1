package helper

import (
	"encoding/json"
	"net/http"
)

func DecodeJSONBody(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
