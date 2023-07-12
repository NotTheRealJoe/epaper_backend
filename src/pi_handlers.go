package epaper_backend

import (
	"log"
	"net/http"

	"github.com/skip2/go-qrcode"
)

// AuthQRHandlerFunc handles endpoint for creating a new authorization code and generating a QR code for it
func (h HandlerHolder) AuthQRHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// create and save authorization code
	authCode := h.repo.CreateAuthorization()

	// generate the qr to a temp file
	qrCodeData, err := qrcode.Encode(
		h.config.PublicBasePath+"/?c="+authCode,
		qrcode.Low,
		h.config.EPaperDisplayHeight,
	)
	if err != nil {
		log.Fatal(err)
	}

	// serve the file
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate")
	w.Write(qrCodeData)
}
