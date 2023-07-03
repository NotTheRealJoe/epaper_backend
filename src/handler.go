package epaper_backend

import (
	"log"
	"net/http"

	"github.com/nottherealjoe/epaper_backend/repository"
	"github.com/skip2/go-qrcode"
)

func NewHandler(repo *repository.MysqlRepository, config *Config) Handler {
	return Handler{
		repo:   repo,
		config: config,
	}
}

type Handler struct {
	repo   *repository.MysqlRepository
	config *Config
}

// SampleHandler is a simple handler that returns a test message
func (h Handler) SampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hellorld!"))
}

func (h Handler) RootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

// === Below are all handlers for the pi functionality ===
// AuthQRHandler handles endpoint for creating a new authorization code and generating a QR code for it
func (h Handler) AuthQRHandler(w http.ResponseWriter, r *http.Request) {
	// create and save authorization code
	authCode := h.repo.CreateAuthoriztion()

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
