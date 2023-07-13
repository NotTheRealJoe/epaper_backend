package epaper_backend

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
			return
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

// == API Handlers ==
func (h HandlerHolder) UploadImageHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	// Check cookie
	cookie, err := r.Cookie(AUTH_COOKIE_NAME)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(401)
			return
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	ok, authorization := h.repo.GetAuthorizationByCookie(cookie.Value)
	if !ok {
		// cookie not ok
		w.WriteHeader(403)
		return
	}

	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
		w.WriteHeader(406)
		return
	}

	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var submitDrawingRequest SubmitDrawingRequest
	err = json.Unmarshal(bodyRaw, &submitDrawingRequest)
	if err != nil {
		log.Print(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	imageData, innerContentType, err := decodeDataURL(submitDrawingRequest.ImageDataURL)
	if err != nil {
		log.Print(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if innerContentType != "image/png" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Only PNG (image/png) encoded data can be accpeted."))
		w.WriteHeader(400)
		return
	}

	drawing := Drawing{
		Author:        submitDrawingRequest.Artist,
		Data:          imageData,
		Authorization: authorization.AuthCode,
	}
	createdId, err := h.repo.SaveDrawing(drawing)
	if err != nil {
		log.Print(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.mqttClient.AddDrawing(createdId, drawing.Data)
	h.repo.RemoveAuthorization(authorization.ID)

	// on success, return 201 with empty body
	w.WriteHeader(201)
}

func decodeDataURL(dataURL string) ([]byte, string, error) {
	if !strings.HasPrefix(dataURL, "data:") {
		return []byte{}, "", fmt.Errorf("string not recognized as a valid data URL")
	}

	contentType := dataURL[5:strings.Index(dataURL, ";")]

	dataURLFormat := dataURL[strings.Index(dataURL, ";")+1 : strings.Index(dataURL, ",")]
	if dataURLFormat != "base64" {
		return []byte{}, "", fmt.Errorf("unrecognized data URL format: " + dataURLFormat)
	}

	decoded, err := base64.StdEncoding.DecodeString(dataURL[strings.Index(dataURL, ",")+1:])
	if err != nil {
		return []byte{}, "", fmt.Errorf("%s :: %w", "failed to decode base64 data", err)
	}

	return decoded, contentType, nil
}
