package epaper_backend

import (
	"errors"
	"log"
	"net/http"
	"strings"
)

const AUTH_COOKIE_NAME = "sauce_auth"

func (h HandlerHolder) RootHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// if they already have a valid userCookie, just serve the page
	userCookie, err := r.Cookie(AUTH_COOKIE_NAME)
	if err == nil {
		// cookie is present
		if h.repo.CookieIsValid(userCookie.Value) {
			http.ServeFile(w, r, h.config.TemplatesPath+"/index.html")
			return
		}
	} else {
		// cookie is not present
		if !errors.Is(err, http.ErrNoCookie) {
			// if we didn't match the no cookie error, there was some weird internal error
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		} // else just continue below
	}

	if !r.URL.Query().Has("c") {
		w.WriteHeader(401)
		w.Write([]byte("Sorry, an auth code is required."))
		return
	}

	authOk, newCookie := h.repo.UseAuthorization(r.URL.Query().Get("c"))
	if !authOk {
		w.WriteHeader(403)
		w.Write([]byte("That auth code is not valid or has expired. Try scanning a fresh QR code."))
		return
	}

	// Tell pi to update the QR code
	h.mqttClient.UpdateQRCode(h.repo.CreateAuthorization())

	// Set the newly generated auth cookie on the client
	w.Header().Add("Set-Cookie", AUTH_COOKIE_NAME+"="+*newCookie)

	http.ServeFile(w, r, h.config.TemplatesPath+"/index.html")
}

func (h HandlerHolder) StaticContentHandlerFunc(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.config.StaticContentPath+"/"+strings.TrimPrefix(r.URL.Path, "/static/"))
}
