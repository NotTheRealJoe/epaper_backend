package epaper_backend

import (
	"net/http"
	"strings"
)

func (h HandlerHolder) RootHandlerFunc(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func (h HandlerHolder) StaticContentHandlerFunc(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.config.StaticContentPath+"/"+strings.TrimPrefix(r.URL.Path, "/static/"))
}
