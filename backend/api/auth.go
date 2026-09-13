package api

import (
	"log"
	"fmt"
	"bytes"
	"io"
	"time"
	"strconv"
	"crypto/rand"
	"net/http"
	"encoding/json"
	"encoding/base64"
	"mud/core"
)

type rcTokenRequestBody struct {
	GrantType string `json:"grant_type"`
	Code string `json:"code"`
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUri string `json:"redirect_uri"`
}

type rcTokenResponseBody struct {
	AccessToken string `json:"access_token"`
}

type rcGetProfilesResponseBody struct {
	Id int `json:"id"`
}

const MUD_SESSION_COOKIE_NAME = "mud_session"

func (apiState *ApiState) HandleAuthLogin(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked /api/auth/login")
	env := core.GetEnv()

	// TODO: state for CSRF
	// Generate state for CSRF prevention
	// stateBytes := make([]byte, 16)
	// rand.Read(stateBytes)

	hostname := request.URL.Query().Get("hostname")
	if hostname == "" {
		http.Error(writer, "No hostname provided.", http.StatusBadRequest)
		return
	}

	redirectUrl := fmt.Sprintf("http://%s:%d/api/auth/callback", hostname, 5173)
	log.Printf("Redirect URL %s", redirectUrl)

	url := fmt.Sprintf(
		"%s?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
		env.RC_AUTH_URL,
		env.RC_CLIENT_ID,
		redirectUrl,
		redirectUrl)

	http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
}

func (apiState *ApiState) HandleAuthCallback(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked /api/auth/callback")
	env := core.GetEnv()

	// Get auth code from URL
	authCode := request.URL.Query().Get("code")
	if authCode == "" {
		http.Error(writer, "Missing authorization code.", http.StatusBadRequest)
		return
	}

	// Get redirect URL from URL
	state := request.FormValue("state")
	log.Printf("Callback redirect URL %s", state)

	// Make token request body
	tokenRequestBody, error := json.Marshal(rcTokenRequestBody {
		GrantType: "authorization_code",
		ClientId: env.RC_CLIENT_ID,
		ClientSecret: env.RC_CLIENT_SECRET,
		Code: authCode,
		RedirectUri: state,
	})
	if error != nil {
		http.Error(writer, error.Error(), http.StatusInternalServerError)
		return
	}

	// Make token request to RC API
	tokenResponse, error := http.Post(env.RC_TOKEN_URL, "application/json", bytes.NewBuffer(tokenRequestBody))
	if error != nil {
		http.Error(writer, error.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		io.Copy(io.Discard, tokenResponse.Body)
		tokenResponse.Body.Close()
	}()

	// Get token response bytes
	tokenResponseBytes, error := io.ReadAll(tokenResponse.Body)
	if error != nil {
		http.Error(writer, error.Error(), http.StatusInternalServerError)
		return
	}

	// Parse token response bytes into body
	var tokenResponseBody rcTokenResponseBody
	error = json.Unmarshal(tokenResponseBytes, &tokenResponseBody)
	if error != nil {
		http.Error(writer, error.Error(), http.StatusInternalServerError)
		return
	}

	// Get user ID from RC API
	userId, error := getUserIdFromRcApi(tokenResponseBody.AccessToken)
	if error != nil {
		http.Error(writer, error.Error(), http.StatusInternalServerError)
		return
	}

	apiState.createSessionForUser(writer, request, userId)
}

func (apiState *ApiState) HandleDebugLogin(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked /api/auth/debug")

	// Get user ID from query params
	userIdString := request.URL.Query().Get("user")
	if userIdString == "" {
		http.Error(writer, "Missing debug user ID.", http.StatusBadRequest)
		return
	}

	// Convert user ID to int
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	apiState.createSessionForUser(writer, request, userId)
}

func getUserIdFromRcApi(token string) (int, error) {
	env := core.GetEnv()

	// Build a GET request to the RC API
	recurseRequest, err := http.NewRequest("GET", fmt.Sprintf("%s/profiles/me", env.RC_API_URL), nil)
	if err != nil {
		return -1, err
	}

	recurseRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Make the GET request
	httpClient := &http.Client{}
	recurseResponse, err := httpClient.Do(recurseRequest)
	if err != nil {
		return -1, err
	}
	defer func() {
		io.Copy(io.Discard, recurseResponse.Body)
		recurseResponse.Body.Close()
	}()

	// For some reason RC returns 404 when you don't provide an auth token
	if recurseResponse.StatusCode == http.StatusNotFound || recurseResponse.StatusCode == http.StatusUnauthorized {
		return -1, nil
	}

	// Get the contents of the response body
	recurseResponseBodyBytes, err := io.ReadAll(recurseResponse.Body)
	if err != nil {
		return -1, err
	}

	// Convert the response from JSON
	var recurseResponseBody rcGetProfilesResponseBody
	err = json.Unmarshal(recurseResponseBodyBytes, &recurseResponseBody)
	if err != nil {
		return -1, err
	}

	return recurseResponseBody.Id, nil
}

func (apiState *ApiState) createSessionForUser(writer http.ResponseWriter, request *http.Request, userId int) {
	// Generate a secure session ID
	sessionTokenBytes := make([]byte, 32)
	rand.Read(sessionTokenBytes)
	sessionToken := base64.URLEncoding.EncodeToString(sessionTokenBytes)

	// Save session state
	apiState.tokenToIdMutex.Lock()
	apiState.tokenToIdMap[sessionToken] = userId
	apiState.tokenToIdMutex.Unlock()

	// Store session state in user's browser
	http.SetCookie(writer, &http.Cookie {
		Name: MUD_SESSION_COOKIE_NAME,
		Value: sessionToken,
		Path: "/",
		Expires: time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure: false, // Set to true in prod to encrypt w/ https
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect user back to home page
	http.Redirect(writer, request, "http://localhost:5173/game", http.StatusSeeOther)
}
