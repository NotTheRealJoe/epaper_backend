package epaper_backend

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/skip2/go-qrcode"
)

type MQTTClient struct {
	client mqtt.Client
	repo   *MysqlRepository
	config *Config
}

func NewMQTTClient(repo *MysqlRepository, config *Config) MQTTClient {
	// set up basic mqtt client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", config.MQTT.BrokerAddress, config.MQTT.BrokerPort))
	opts.SetClientID(config.MQTT.Prefix + strconv.Itoa(rand.Intn(999)))
	opts.SetUsername(config.MQTT.Username)
	opts.SetPassword(config.MQTT.Password)

	// attach TLS config to the client options
	opts.SetTLSConfig(&tls.Config{
		ServerName: config.MQTT.BrokerTLSHostname,
	})

	// construct client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(fmt.Errorf("%s :: %w", "Failed to connect to MQTT server.", token.Error()))
	}
	log.Println("MQTT connection established")

	mq := MQTTClient{
		client: client,
		repo:   repo,
		config: config,
	}

	// Add subscriptions
	mq.client.Subscribe("epaper/online", 2, mq.ReceiveStartup)

	mq.repo = repo
	mq.config = config

	return mq
}

// === Subscription handler functions

func (m MQTTClient) ReceiveStartup(client mqtt.Client, msg mqtt.Message) {
	newAuthorization := m.repo.CreateAuthorization()
	m.UpdateQRCode(newAuthorization)
}

// === Publishing functions ===

func (m MQTTClient) UpdateQRCode(newAuthCode string) {
	generated_code, err := qrcode.Encode(m.config.PublicBasePath+"/?c="+newAuthCode, qrcode.Low, 122)
	if err != nil {
		log.Fatal(err)
	}

	m.client.Publish("epaper/cmnd/update-qr", 2, false, generated_code)
}

func (m MQTTClient) AddDrawing(id int64, data []byte) {
	m.client.Publish(fmt.Sprintf("epaper/cmnd/image/add/%d", id), 2, false, data)
}

func (m MQTTClient) RemoveDrawing(id int) {
	m.client.Publish("epaper/cmnd/image/remove", 2, false, strconv.Itoa(id))
}

func (m MQTTClient) BlankScreen() {
	m.client.Publish("epaper/cmnd/blank", 2, false, "true")
}

func (m MQTTClient) UnblankScreen() {
	m.client.Publish("epaper/cmnd/unblank", 2, false, "false")
}
