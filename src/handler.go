package epaper_backend

import (
	"net/http"
)

func NewHandlerHolder(repo *MysqlRepository, config *Config, mqttClient *MQTTClient) HandlerHolder {
	return HandlerHolder{
		repo:       repo,
		config:     config,
		mqttClient: mqttClient,
	}
}

type HandlerHolder struct {
	repo       *MysqlRepository
	config     *Config
	mqttClient *MQTTClient
}

// SampleHandler is a simple handler that returns a test message
func (h HandlerHolder) SampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hellorld!"))
}
