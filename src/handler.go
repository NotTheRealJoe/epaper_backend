package epaper_backend

import (
	"net/http"

	"github.com/nottherealjoe/epaper_backend/repository"
)

func NewHandler(repo *repository.MysqlRepository, config *Config) HandlerHolder {
	return HandlerHolder{
		repo:   repo,
		config: config,
	}
}

type HandlerHolder struct {
	repo   *repository.MysqlRepository
	config *Config
}

// SampleHandler is a simple handler that returns a test message
func (h HandlerHolder) SampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hellorld!"))
}
