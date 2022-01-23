package responses

import (
	"encoding/json"
	"net/http"
)

func NewJsonResponse(response http.ResponseWriter, status int, body interface{}) {
	response.WriteHeader(status)
	json.NewEncoder(response).Encode(body)
	response.Header().Add("content-type", "application/json")
}
