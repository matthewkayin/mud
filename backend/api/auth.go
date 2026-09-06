package api

import (
	"log"
	"bytes"
	"io"
	"net/http"
	"encoding/json"
	"mud/mudenv"
)

type PostAuthRequest struct {
	Code string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectUri string `json:"redirect_uri"`
}

type RecursePostTokenBody struct {
	GrantType string `json:"grant_type"`
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectUri string `json:"redirect_uri"`
}

func HandlePostAuth(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked POST /api/auth")
	env := mudenv.Get()

	// Get request data into struct
	var requestData PostAuthRequest
	parseError := json.NewDecoder(request.Body).Decode(&requestData)
	if parseError != nil {
		http.Error(writer, parseError.Error(), http.StatusBadRequest)
		return
	}

	// Create body for request to RC API
	recurseRequestBody, marshalError := json.Marshal(RecursePostTokenBody {
		GrantType: "authorization_code",
		ClientId: env.RecurseClientId,
		ClientSecret: env.RecurseClientSecret,
		Code: requestData.Code,
		CodeVerifier: requestData.CodeVerifier,
		RedirectUri: requestData.RedirectUri,
	})
	if marshalError != nil {
		http.Error(writer, marshalError.Error(), http.StatusInternalServerError)
		return
	}

	// Make request to RC API
	recurseResponse, recurseError := http.Post(env.RecurseTokenUrl, "application/json", bytes.NewBuffer(recurseRequestBody))
	if recurseError != nil {
		http.Error(writer, recurseError.Error(), http.StatusInternalServerError)
		return
	}

	// Pass the recurse response as the response to this request
	defer recurseResponse.Body.Close()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, copyError := io.Copy(writer, recurseResponse.Body)
	if copyError != nil {
		http.Error(writer, copyError.Error(), http.StatusInternalServerError)
		return
	}
}
