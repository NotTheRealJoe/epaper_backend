package epaper_backend

import (
	"net/http"
	"strings"
)

const AUTH_COOKIE_NAME = "sauce_auth"

func (h HandlerHolder) RootHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("c") {
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized - an auth code is required."))
		return
	}

	authOk, cookie := h.repo.UseAuthorization(r.URL.Query().Get("c"))
	if !authOk {
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized - an auth code is required."))
		return
	}

	w.Header().Add("Set-Cookie", AUTH_COOKIE_NAME+"="+*cookie)
	//TODO: Tell pi to update the QR code

	http.ServeFile(w, r, h.config.StaticContentPath+"/index.html")
}

func (h HandlerHolder) StaticContentHandlerFunc(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.config.StaticContentPath+"/"+strings.TrimPrefix(r.URL.Path, "/static/"))
}
