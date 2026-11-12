package api

import (
	"log"
	"errors"
	"strings"
	"net/http"
	"crypto/rand"
	"encoding/json"
	"mud/env"
)

type authCredentialsRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authUserResponseBody struct {
	Username string `json:"username"`
}

type authErrorResponseBody struct {
	Error string `json:"error"`
}

type authConfigResponseBody struct {
	DebugAuth bool `json:"debugAuth"`
}

type authDebugLoginRequestBody struct {
	Username string `json:"username"`
}

const MUD_SESSION_COOKIE_NAME = "mud_session"
const AUTH_REQUEST_BODY_MAX_BYTES = 4096

func (apiState *ApiState) RegisterAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/me", apiState.HandleAuthMe)
	mux.HandleFunc("POST /api/auth/register", apiState.HandleAuthRegister)
	mux.HandleFunc("POST /api/auth/login", apiState.HandleAuthLogin)
	mux.HandleFunc("POST /api/auth/logout", apiState.HandleAuthLogout)
	mux.HandleFunc("GET /api/auth/config", apiState.HandleAuthConfig)
	if env.Get().ENABLE_DEBUG_AUTH {
		mux.HandleFunc("POST /api/auth/debug-login", apiState.HandleAuthDebugLogin)
	}
}

func (apiState *ApiState) HandleAuthConfig(writer http.ResponseWriter, request *http.Request) {
	writeJson(writer, http.StatusOK, authConfigResponseBody { DebugAuth: env.Get().ENABLE_DEBUG_AUTH })
}

// Only checks that the username is valid. Debug logins have no session or cookie:
// the client keeps the username itself and sends it when opening the websocket.
func (apiState *ApiState) HandleAuthDebugLogin(writer http.ResponseWriter, request *http.Request) {
	var requestBody authDebugLoginRequestBody
	request.Body = http.MaxBytesReader(writer, request.Body, AUTH_REQUEST_BODY_MAX_BYTES)
	err := json.NewDecoder(request.Body).Decode(&requestBody)
	if err != nil {
		writeAuthError(writer, http.StatusBadRequest, "The request was not valid.")
		return
	}

	err = validateUsername(requestBody.Username)
	if err != nil {
		writeAuthError(writer, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(writer, http.StatusOK, authUserResponseBody { Username: requestBody.Username })
}

func (apiState *ApiState) HandleAuthMe(writer http.ResponseWriter, request *http.Request) {
	username, authenticated := apiState.getSessionUsername(request)
	if !authenticated {
		writeAuthError(writer, http.StatusUnauthorized, "You are not signed in.")
		return
	}

	writeJson(writer, http.StatusOK, authUserResponseBody { Username: username })
}

func (apiState *ApiState) HandleAuthRegister(writer http.ResponseWriter, request *http.Request) {
	credentials, err := readCredentials(writer, request)
	if err != nil {
		writeAuthError(writer, http.StatusBadRequest, "The request was not valid.")
		return
	}

	account, err := createAccount(credentials.Username, credentials.Password)
	var validationError *accountValidationError
	if errors.As(err, &validationError) {
		writeAuthError(writer, http.StatusBadRequest, validationError.message)
		return
	}
	if errors.Is(err, errAccountUsernameTaken) {
		writeAuthError(writer, http.StatusConflict, "That username is taken.")
		return
	}
	if err != nil {
		log.Printf("Warn - failed to create account %s: %s", credentials.Username, err.Error())
		writeAuthError(writer, http.StatusInternalServerError, "Could not create the account.")
		return
	}

	log.Printf("Created account %s.", account.Username)
	apiState.createSession(writer, request, account.Username)
	writeJson(writer, http.StatusOK, authUserResponseBody { Username: account.Username })
}

func (apiState *ApiState) HandleAuthLogin(writer http.ResponseWriter, request *http.Request) {
	credentials, err := readCredentials(writer, request)
	if err != nil {
		writeAuthError(writer, http.StatusBadRequest, "The request was not valid.")
		return
	}

	account, err := loadAccount(credentials.Username)
	if errors.Is(err, errAccountNotFound) {
		writeAuthError(writer, http.StatusUnauthorized, "Incorrect username or password.")
		return
	}
	if err != nil {
		writeAuthError(writer, http.StatusInternalServerError, "Could not sign in.")
		return
	}

	passwordMatches := checkAccountPassword(account, credentials.Password)
	if !passwordMatches {
		writeAuthError(writer, http.StatusUnauthorized, "Incorrect username or password.")
		return
	}

	// Respond with the stored username so the display capitalization doesn't depend on how it was typed at login
	apiState.createSession(writer, request, account.Username)
	writeJson(writer, http.StatusOK, authUserResponseBody { Username: account.Username })
}

func (apiState *ApiState) HandleAuthLogout(writer http.ResponseWriter, request *http.Request) {
	sessionCookie, err := request.Cookie(MUD_SESSION_COOKIE_NAME)
	if err == nil {
		apiState.sessionMutex.Lock()
		delete(apiState.sessionTokenToUsername, sessionCookie.Value)
		apiState.sessionMutex.Unlock()
	}

	http.SetCookie(writer, &http.Cookie {
		Name: MUD_SESSION_COOKIE_NAME,
		Value: "",
		Path: "/",
		MaxAge: -1,
		HttpOnly: true,
		Secure: isSecureRequest(request),
		SameSite: http.SameSiteLaxMode,
	})
	writer.WriteHeader(http.StatusNoContent)
}

func (apiState *ApiState) createSession(writer http.ResponseWriter, request *http.Request, username string) {
	sessionToken := rand.Text()

	apiState.sessionMutex.Lock()
	apiState.sessionTokenToUsername[sessionToken] = username
	apiState.sessionMutex.Unlock()

	// No expiry, so the cookie lasts as long as the browser session.
	// The server forgets sessions when it restarts anyway.
	http.SetCookie(writer, &http.Cookie {
		Name: MUD_SESSION_COOKIE_NAME,
		Value: sessionToken,
		Path: "/",
		HttpOnly: true,
		Secure: isSecureRequest(request),
		SameSite: http.SameSiteLaxMode,
	})
}

// Returns the signed-in account's username in its display form
func (apiState *ApiState) getSessionUsername(request *http.Request) (string, bool) {
	sessionCookie, err := request.Cookie(MUD_SESSION_COOKIE_NAME)
	if err != nil {
		return "", false
	}

	apiState.sessionMutex.RLock()
	username, authenticated := apiState.sessionTokenToUsername[sessionCookie.Value]
	apiState.sessionMutex.RUnlock()

	return username, authenticated
}

func (apiState *ApiState) getSessionKey(request *http.Request) (string, bool) {
	username, ok := apiState.getSessionUsername(request)
	return strings.ToLower(username), ok
}

// True when the client reached the server over HTTPS, either directly or through a proxy
func isSecureRequest(request *http.Request) bool {
	return request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https"
}

func readCredentials(writer http.ResponseWriter, request *http.Request) (authCredentialsRequestBody, error) {
	var credentials authCredentialsRequestBody
	request.Body = http.MaxBytesReader(writer, request.Body, AUTH_REQUEST_BODY_MAX_BYTES)
	err := json.NewDecoder(request.Body).Decode(&credentials)
	return credentials, err
}

func writeAuthError(writer http.ResponseWriter, status int, message string) {
	writeJson(writer, status, authErrorResponseBody { Error: message })
}

func writeJson(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	err := json.NewEncoder(writer).Encode(value)
	if err != nil {
		log.Printf("Warn - failed to write JSON response: %s", err.Error())
	}
}
