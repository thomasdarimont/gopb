package main

import (
	"fmt"
	"net/http"
	"strings"

	"crypto/md5"
	"io"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"

	"github.com/prometheus/common/log"
	gominiauthcommon "github.com/stretchr/gomniauth/common"
)

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}

type chatUser struct {
	gominiauthcommon.User
	uniqueID string
}

func (u chatUser) UniqueID() string {
	return u.uniqueID
}

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("auth")

	if err == http.ErrNoCookie || cookie.Value == "" {
		// not authenticated
		w.Header().Set("location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// success cann the next handler
	h.next.ServeHTTP(w, r)
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// loginHandler handles third-party login process
// format: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {

	segs := strings.Split(r.URL.Path, "/")
	if len(segs) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Path must be /auth/{action}/{provider}")
	}
	action := segs[2]
	provider := segs[3]

	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
			return
		}
		loginURL, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to GetBeginAuthURL for %s: %s", provider, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("location", loginURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(w, fmt.Sprintf("error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
			return
		}

		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			http.Error(w, fmt.Sprintf("error when trying to complete auth for %s: %s", provider, err), http.StatusInternalServerError)
			return
		}
		user, err := provider.GetUser(creds)
		if err != nil {
			http.Error(w, fmt.Sprintf("error when trying to get user for %s: %s", provider, err), http.StatusInternalServerError)
			return
		}

		chatUser := &chatUser{User: user}
		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Email()))
		chatUser.uniqueID = fmt.Sprintf("%x", m.Sum(nil))
		avatarURL, err := avatars.GetAvatarURL(chatUser)

		if err != nil {
			log.Fatalln("Error when trying to GetAuthURL", "-", err)
		}

		authCookieValue := objx.New(map[string]interface{}{
			"userId":     chatUser.uniqueID,
			"name":       user.Name(),
			"avatar_url": avatarURL,
		}).MustBase64()

		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/",
		})

		w.Header().Set("location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
