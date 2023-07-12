package epaper_backend

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/skip2/go-qrcode"
)

type MQTTClient struct {
	client mqtt.Client
	repo   MysqlRepository
	config Config
}

func NewMQTTClient(repo MysqlRepository, config Config) MQTTClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.MQTTBrokerAddress, config.MQTTBrokerPort))
	opts.SetClientID(config.MQTTPrefix + strconv.Itoa(rand.Intn(999)))
	opts.SetUsername(config.MQTTUsername)
	opts.SetPassword(config.MQTTPassword)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	mq := MQTTClient{
		client: client,
	}

	mq.client.Subscribe("epaper/online", 2, mq.ReceiveStartup)

	mq.repo = repo
	mq.config = config

	return mq
}

// === Subscription handler functions

func (m MQTTClient) ReceiveStartup(client mqtt.Client, msg mqtt.Message) {
	newAuthorization := m.repo.CreateAuthoriztion()
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

func (m MQTTClient) AddDrawing(id int, data []byte) {
	m.client.Publish("epaper/cmnd/image/add/"+strconv.Itoa(id), 2, false, data)
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
