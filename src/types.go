package epaper_backend

type Config struct {
	ListenPort          int    `json:"listenPort"`
	DBHost              string `json:"dbHost"`
	DBPort              string `json:"dbPort"`
	DBUsername          string `json:"dbUsername"`
	DBPassword          string `json:"dbPassword"`
	DBName              string `json:"dbName"`
	PublicBasePath      string `json:"publicBasePath"`
	EPaperDisplayHeight int    `json:"ePaperDisplayHeight"`
	StaticContentPath   string `json:"staticContentPath"`
	MQTTBrokerAddress   string `json:"mqttBrokerAddress"`
	MQTTBrokerPort      int    `json:"mqttBrokerPort"`
	MQTTUsername        string `json:"mqttUsername"`
	MQTTPassword        string `json:"mqttPassword"`
	MQTTPrefix          string `json:"mqttPrefix"`
}

type Authorzation struct {
	ID         int
	AuthCode   string
	UserCookie *string
}
