package epaper_backend

import (
	"net/http"
	"strings"
)
func (h HandlerHolder) StaticContentHandlerFunc(w http.ResponseWriter, r *http.Request) {
	println("staticcontenthandler called")
	file := strings.TrimPrefix(r.URL.Path, "/static/")
	println(h.config.StaticContentPath + "/" + file)
	http.ServeFile(w, r, h.config.StaticContentPath+"/"+file)
}
