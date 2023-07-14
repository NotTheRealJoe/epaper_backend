package epaper_backend

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

func (h HandlerHolder) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=172800")
	http.ServeFile(w, r, h.config.StaticContentPath+"/favicon.ico")
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

func (h HandlerHolder) AdminGetDrawingsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if !h.verifyAdminBasicAuth(r) {
		w.WriteHeader(403)
		return
	}

	drawings, err := h.repo.GetAllDrawingsRemovedLast()
	if err != nil {
		formattedError := fmt.Sprintf("%s:\n%v", "Failed to get drawings from database:", err)
		log.Print(formattedError)
		w.WriteHeader(500)
		w.Write([]byte(formattedError))
	}

	encoded, _ := json.Marshal(*drawings)
	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

func (h HandlerHolder) AdminGetDrawingDataHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if !h.verifyAdminBasicAuth(r) {
		w.WriteHeader(403)
		return
	}

	drawingIDStr := strings.TrimPrefix(r.URL.Path, "/admin/api/drawing/")
	drawingID, err := strconv.Atoi(drawingIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(400)
		w.Write([]byte("unable to parse given id parsed as an integer"))
		return
	}

	data, err := h.repo.GetDrawingData(drawingID)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(*data)
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

func (h HandlerHolder) verifyAdminBasicAuth(r *http.Request) bool {
	authHeader := r.Header.Get("authorization")
	if authHeader == "" {
		return false
	}

	// for safety, disallow if config username or password are empty
	if h.config.Admin.Username == "" || h.config.Admin.PasswordHashBase64 == "" {
		return false
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Basic" {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(headerParts[1])
	if err != nil {
		return false
	}

	decodedParts := strings.Split(string(decoded), ":")
	//return len(decodedParts) == 2 && headerParts[0] == h.config.Admin.Username && sha256.New(headerParts[1]) != config.Admin.PasswordHash
	if len(decodedParts) != 2 || decodedParts[0] != h.config.Admin.Username {
		return false
	}

	return byteSlicesEqual(
		passwordHash(decodedParts[1]),
		h.config.Admin.PasswordHash,
	)
}

func byteSlicesEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func passwordHash(password string) []byte {
	h := sha256.New()
	h.Write([]byte(password))
	return h.Sum(nil)
}
